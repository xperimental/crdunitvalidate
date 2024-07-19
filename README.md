# crdunitvalidate

Library to easily test validations present in a Kubernetes CustomResourceDefinition without running a Kubernetes control-plane.

**Note:** This is still very much a work-in-progress, including the name.

## Usage

This library can be used to test validations in custom resources using a unit test instead of running a Kubernetes control-plane, loading the CustomResourceDefinition into the API and making requests to the api-server to test the validations.

The validations are based on the CustomResourceDefinition itself and the code present in the Kubernetes API-server, so there's no need to have any Go definitions of the custom resources.

```golang
// Load a resource from YAML
bytes, err := os.ReadFile(pathToResourceFile)
res, err := crdunitvalidate.LoadYAML(bytes)

// Instantiate validator using CRD
v, err := crdunitvalidate.NewValidator(pathToCRDFile)

// Use validator to test validations
// Validation failures are returned as errors
err := v.Validate(res)
```

See also the examples in `validate_test.go` and `testdata`.

