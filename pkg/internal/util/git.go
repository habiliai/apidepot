package util

import (
	"archive/tar"
	"github.com/emirpasic/gods/stacks/arraystack"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"
	"io"
	"path/filepath"
)

type WalkGitTreeFunc func(
	basePath string,
	entry object.TreeEntry,
	fileContents []byte,
) error

func WalkGitTree(tree *object.Tree, cb WalkGitTreeFunc) error {
	type StackEntry struct {
		tree     *object.Tree
		basePath string
	}
	trees := arraystack.New()
	trees.Push(StackEntry{tree, ""})

	for !trees.Empty() {
		v, ok := trees.Pop()
		if !ok {
			return errors.New("failed to pop stack entry")
		}
		parent := v.(StackEntry)

		for _, entry := range parent.tree.Entries {
			var fileContents []byte
			if !entry.Mode.IsFile() {
				subTree, err := parent.tree.Tree(entry.Name)
				if err != nil {
					return errors.Wrapf(err, "failed to get sub tree")
				}

				trees.Push(StackEntry{subTree, filepath.Join(parent.basePath, entry.Name)})
			} else {
				file, err := parent.tree.TreeEntryFile(&entry)
				if err != nil {
					return errors.Wrapf(err, "failed to get tree entry file")
				}

				reader, err := file.Reader()
				if err != nil {
					return errors.Wrapf(err, "failed to get reader")
				}

				fileContents, err = io.ReadAll(reader)
				if err != nil {
					return errors.Wrapf(err, "failed to read all")
				}
			}

			if err := cb(parent.basePath, entry, fileContents); err != nil {
				return errors.Wrapf(err, "failed to walk tree")
			}
		}
	}

	return nil
}

func ArchiveGitTree(
	tree *object.Tree,
	wr io.Writer,
) (err error) {
	tw := tar.NewWriter(wr)
	defer tw.Close()

	if err := WalkGitTree(tree, func(basePath string, file object.TreeEntry, contents []byte) error {
		logger.Debug("receive", "file", file.Name)

		osFileMode, err := file.Mode.ToOSFileMode()
		if err != nil {
			return errors.Wrapf(err, "failed to convert file mode")
		}
		header := tar.Header{
			Name: filepath.Join(basePath, file.Name),
			Mode: int64(osFileMode),
			Size: int64(len(contents)),
		}

		switch file.Mode {
		case filemode.Dir:
			header.Typeflag = tar.TypeDir
		case filemode.Regular, filemode.Executable, filemode.Empty, filemode.Deprecated:
			header.Typeflag = tar.TypeReg
		case filemode.Symlink:
			header.Typeflag = tar.TypeSymlink
		default:
			return errors.Errorf("unsupported file mode: %s", file.Mode)
		}

		if file.Name == "" {
			return errors.Errorf("file name is empty")
		}

		// write header
		if err := tw.WriteHeader(&header); err != nil {
			return err
		}

		// if not a dir, write file content
		if contents != nil {
			if _, err := tw.Write(contents); err != nil {
				return errors.Wrapf(err, "failed to write file content")
			}
		}

		return nil
	}); err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return errors.Wrapf(err, "failed to close tar writer")
	}

	return nil
}
