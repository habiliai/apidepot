package instance

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/habiliai/apidepot/pkg/internal/constants"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strings"
	"time"
)

func (s *service) isAvailable(
	ctx context.Context,
	instance *domain.Instance,
	readTimeout time.Duration,
) (map[string]bool, error) {
	type Result struct {
		name string
		ok   bool
	}

	endpoint := instance.Stack.Endpoint()

	var checkRequests []map[string]string
	if instance.Stack.AuthEnabled {
		checkRequests = append(checkRequests, map[string]string{
			"name": "auth",
			"path": endpoint + constants.PathAuthHealth,
		})
	}

	if instance.Stack.StorageEnabled {
		checkRequests = append(checkRequests, map[string]string{
			"name": "storage",
			"path": endpoint + constants.PathStorageHealth,
		})
	}

	if instance.Stack.PostgrestEnabled {
		checkRequests = append(checkRequests, map[string]string{
			"name": "postgrest-live",
			"path": endpoint + constants.PathPostgrestLive,
		})
		checkRequests = append(checkRequests, map[string]string{
			"name": "postgrest-ready",
			"path": endpoint + constants.PathPostgrestReady,
		})
	}

	rels, err := s.vapis.GetAllDependenciesOfVapiReleases(ctx, instance.Stack.GetVapiReleases())
	if err != nil {
		return nil, err
	}

	logger.Debug("print", "len(checkRequests)", len(checkRequests), "len(rels)", len(rels))

	checksCh := make(chan Result, len(checkRequests)+len(rels))
	var eg errgroup.Group
	for _, checkReq := range checkRequests {
		eg.Go(func() error {
			result := Result{
				name: checkReq["name"],
				ok:   false,
			}
			if err := func() error {
				ctx, cancel := context.WithTimeout(ctx, readTimeout)
				defer cancel()

				req, err := http.NewRequestWithContext(
					ctx,
					http.MethodGet,
					checkReq["path"],
					nil,
				)
				req.Header.Add("Authorization", "Bearer "+instance.Stack.AnonApiKey)
				if err != nil {
					return errors.Wrapf(err, "failed to create request")
				}

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "i/o timeout") {
						return nil
					}

					if e := (*tls.CertificateVerificationError)(nil); errors.As(err, &e) {
						logger.Warn(fmt.Sprintf("'%s' is unavailable. because of %v", result.name, e.Unwrap()))
						return nil
					}
					return errors.Wrapf(err, "failed to do request")
				}

				result.ok = resp.StatusCode == http.StatusOK
				return nil
			}(); err != nil {
				return err
			}

			select {
			case <-ctx.Done():
				return errors.WithStack(ctx.Err())
			case checksCh <- result:
				return nil
			}
		})
	}

	for _, rel := range rels {
		eg.Go(func() error {
			result := Result{
				name: rel.Package.Name,
				ok:   false,
			}
			if err := func() error {
				ctx, cancel := context.WithTimeout(ctx, readTimeout)
				defer cancel()

				req, err := http.NewRequestWithContext(
					ctx,
					http.MethodGet,
					endpoint+constants.PathVapis+"/"+rel.Slug()+constants.VapiHealthPath,
					nil,
				)
				req.Header.Add("Authorization", "Bearer "+instance.Stack.AnonApiKey)
				if err != nil {
					return errors.Wrapf(err, "failed to create request")
				}

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					if errors.Is(err, context.DeadlineExceeded) {
						return nil
					}

					if e := (*tls.CertificateVerificationError)(nil); errors.As(err, &e) {
						logger.Warn(fmt.Sprintf("'%s' is unavailable. because of %v", rel.Package.Name, e.Unwrap()))
						return nil
					}
					return errors.Wrapf(err, "failed to do request")
				}

				result.ok = resp.StatusCode == http.StatusOK
				return nil
			}(); err != nil {
				return err
			}

			select {
			case <-ctx.Done():
				return errors.WithStack(ctx.Err())
			case checksCh <- result:
				return nil
			}
		})
	}

	go func() {
		err = eg.Wait()
		close(checksCh)
	}()

	results := map[string]bool{}
	for interrupt := false; !interrupt; {
		select {
		case <-ctx.Done():
			return nil, errors.WithStack(ctx.Err())
		case result, ok := <-checksCh:
			if !ok {
				interrupt = true
				break
			}
			if !result.ok {
				logger.Warn("unavailable", "name", result.name)
			}
			results[result.name] = result.ok
		}
	}

	logger.Debug("print", "results", results)

	if err != nil {
		return nil, err
	}

	return results, nil
}
