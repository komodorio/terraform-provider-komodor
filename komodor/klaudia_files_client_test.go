package komodor

import (
	"io"
	"mime"
	"mime/multipart"
	"testing"
)

func TestBuildKlaudiaFileMultipartBody(t *testing.T) {
	body, contentType, err := buildKlaudiaFileMultipartBody(
		[]klaudiaFilePayload{{Filename: "runbook.md", Content: []byte("hello")}},
		"files",
		&KlaudiaFileClusters{Include: []string{"prod"}, Exclude: []string{"dev"}},
		true,
	)
	if err != nil {
		t.Fatal(err)
	}

	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatal(err)
	}
	reader := multipart.NewReader(body, params["boundary"])

	parts := map[string]string{}
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		content, err := io.ReadAll(part)
		if err != nil {
			t.Fatal(err)
		}
		parts[part.FormName()] = string(content)
		if part.FormName() == "files" && part.FileName() != "runbook.md" {
			t.Fatalf("unexpected filename %q", part.FileName())
		}
	}

	if parts["files"] != "hello" {
		t.Fatalf("unexpected file content %q", parts["files"])
	}
	if parts["clusters"] != `[{"include":["prod"],"exclude":["dev"]}]` {
		t.Fatalf("unexpected clusters payload %q", parts["clusters"])
	}
}

func TestBuildKlaudiaFileMultipartBodyWithUpdateFileField(t *testing.T) {
	body, contentType, err := buildKlaudiaFileMultipartBody(
		[]klaudiaFilePayload{{Filename: "runbook.md", Content: []byte("hello")}},
		"file",
		nil,
		false,
	)
	if err != nil {
		t.Fatal(err)
	}

	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatal(err)
	}
	reader := multipart.NewReader(body, params["boundary"])

	part, err := reader.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	if part.FormName() != "file" {
		t.Fatalf("unexpected form field %q", part.FormName())
	}
	if part.FileName() != "runbook.md" {
		t.Fatalf("unexpected filename %q", part.FileName())
	}
}
