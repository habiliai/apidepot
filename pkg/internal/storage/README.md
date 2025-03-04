# Storage GO

This is forked from [supabase/storage-go](https://github.com/supabase-community/storage-go)

This library is a Golang client for the [Supabase Storage API](https://supabase.com/docs/guides/storage). It's a collection of helper functions that help you manage your buckets through the API.

## Quick start guide

#### Install

```shell
go get github.com/supabase-community/storage-go
```

### Connecting to the storage backend

```go
package main

import (
	"fmt"
	"log"
	"os"

	storage_go "github.com/supabase-community/storage-go"
)

func main() {
	storageClient := storage_go.NewClient("https://<project-reference-id>.supabase.co/storage/v1", "<project-secret-api-key>", nil)
}
```

### Handling resources

#### Handling Storage Buckets

- Create a new Storage bucket:

```go
  result, err := storageClient.CreateBucket(ctx, "bucket-id", storage_go.BucketOptions{
    Public: true,
  })
```

- Retrieve the details of an existing Storage bucket:

```go
  result, err := storageClient.GetBucket(ctx, "bucket-id")
```

- Update a new Storage bucket:

```go
  result, err := storageClient.UpdateBucket(ctx, "bucket-id", storage_go.BucketOptions{
    Public: true,
  })
```

- Remove all objects inside a single bucket:

```go
  result, err := storageClient.EmptyBucket(ctx, "bucket-id")
```

- Delete an existing bucket (a bucket can't be deleted with existing objects inside it):

```go
  result, err := storageClient.DeleteBucket(ctx, "bucket-id")
```

- Retrieve the details of all Storage buckets within an existing project:

```go
  result, err := storageClient.ListBuckets(ctx, "bucket-id")
```

#### Handling Files

```go
  fileBody := ... // load your file here

  result, err := storageClient.UploadFile(ctx, "test", "test.txt", fileBody)
```

> Note: The `upload` method also accepts a map of optional parameters.

- Download a file from an exisiting bucket:

```go
  result, err := storageClient.DownloadFile(ctx, "bucket-id", "test.txt")
```

- List all the files within a bucket:

```go
  result, err := storageClient.ListFiles(ctx, "bucket-id", "", storage_go.FileSearchOptions{
      Limit:  10,
      Offset: 0,
      SortByOptions: storage_go.SortBy{
      Column: "",
      Order:  "",
    },
  })
```

> Note: The `list` method also accepts a map of optional parameters.

- Replace an existing file at the specified path with a new one:

```go
  fileBody := ... // load your file here

  result, err := storageClient.UpdateFile(ctx, "test", "test.txt", file)
```

- Move an existing file:

```go
  result, err := storageClient.MoveFile(ctx, "test", "test.txt", "random/test.txt")
```

- Delete files within the same bucket:

```go
  result, err := storageClient.RemoveFile(ctx, "test", []string{"book.pdf"})
```

- Create signed URL to download file without requiring permissions:

```go
  const expireIn = 60

  result, err := storageClient.CreateSignedUrl(ctx, "test", "test.mp4", expireIn)
```

- Retrieve URLs for assets in public buckets:

```go
  result, err := storageClient.GetPublicUrl("test", "book.pdf")
```

- Create an signed URL and upload to signed URL:

```go
  fileBody := ... // load your file here

  resp, err := storageClient.CreateSignedUploadUrl(ctx, "test", "test.txt")
  res, err := storageClient.UploadToSignedUrl(ctx, resp.Url, file)
```

## License

<!-- I don't know which to use, but explicitly stating the license would be a big help -->
