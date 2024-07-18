package crdunitvalidate

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestValidation(t *testing.T) {
	crdFile := "testdata/test-crd.yaml"

	v, err := NewValidator(crdFile)
	if err != nil {
		t.Fatalf("error creating validator: %s", err)
	}

	tt := []struct {
		fileName string
		wantErr  string
	}{
		{
			fileName: "test-valid.yaml",
			wantErr:  "",
		},
		{
			fileName: "test-invalid-required.yaml",
			wantErr: `validation failure list:
spec.image in body is required
spec: Invalid value: "object": no such key: image evaluating rule: image has to use our repository`,
		},
		{
			fileName: "test-invalid-type.yaml",
			wantErr: `validation failure list:
spec.replicas in body must be of type integer: "string"`,
		},
		{
			fileName: "test-invalid-cel.yaml",
			wantErr: `validation failure list:
spec: Invalid value: "object": image has to use our repository`,
		},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.fileName, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join("testdata", tc.fileName)
			bytes, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("error reading file: %s", err)
			}

			resource, err := LoadYAML(bytes)
			if err != nil {
				t.Fatalf("error loading resource: %s", err)
			}

			ctx := context.Background()
			err = v.Validate(ctx, resource)
			errMsg := ""
			if err != nil {
				errMsg = err.Error()
			}

			if errMsg != tc.wantErr {
				t.Errorf("got error %q, want error %q", errMsg, tc.wantErr)
			}
		})
	}
}
