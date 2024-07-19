package crdunitvalidate

import (
	"context"
	"fmt"
	"os"

	apiextensionshelpers "k8s.io/apiextensions-apiserver/pkg/apihelpers"
	apiextensionsinternal "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	structuralschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema/cel"
	apiservervalidation "k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	celconfig "k8s.io/apiserver/pkg/apis/cel"
	"k8s.io/kube-openapi/pkg/validation/errors"
)

type Validator struct {
	scheme *runtime.Scheme
	crd    *apiextensionsv1.CustomResourceDefinition
}

func NewValidator(crdFileName string) (*Validator, error) {
	scheme := runtime.NewScheme()
	if err := apiextensionsv1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("error initializing scheme: %w", err)
	}

	crd, err := loadCRD(scheme, crdFileName)
	if err != nil {
		return nil, fmt.Errorf("error loading CRD: %w", err)
	}

	return &Validator{
		scheme: scheme,
		crd:    crd,
	}, nil
}

func loadCRD(scheme *runtime.Scheme, fileName string) (*apiextensionsv1.CustomResourceDefinition, error) {
	bytes, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	obj, _, err := serializer.NewCodecFactory(scheme).UniversalDeserializer().Decode(bytes, nil, nil)
	if err != nil {
		return nil, err
	}

	return obj.(*apiextensionsv1.CustomResourceDefinition), nil
}

func (v *Validator) Validate(resource *unstructured.Unstructured) error {
	gvk := resource.GetObjectKind().GroupVersionKind()

	if gvk.Group != v.crd.Spec.Group {
		return fmt.Errorf("group differs: %s vs %s", gvk.Group, v.crd.Spec.Group)
	}

	if gvk.Kind != v.crd.Spec.Names.Kind {
		return fmt.Errorf("kind differs: %s vs %s", gvk.Kind, v.crd.Spec.Names.Kind)
	}

	crdSchema, err := apiextensionshelpers.GetSchemaForVersion(v.crd, gvk.Version)
	if err != nil {
		return fmt.Errorf("error getting CRD schema: %w", err)
	}

	internalValidationSchema := &apiextensionsinternal.CustomResourceValidation{}
	if err := apiextensionsv1.Convert_v1_CustomResourceValidation_To_apiextensions_CustomResourceValidation(crdSchema, internalValidationSchema, nil); err != nil {
		return fmt.Errorf("failed to convert CRD validation to internal version: %w", err)
	}

	validator, _, err := apiservervalidation.NewSchemaValidator(internalValidationSchema.OpenAPIV3Schema)
	if err != nil {
		return fmt.Errorf("error creating validator: %w", err)
	}

	ss, err := structuralschema.NewStructural(internalValidationSchema.OpenAPIV3Schema)
	if err != nil {
		return fmt.Errorf("error creating structural schema: %w", err)
	}

	errList := []error{}
	result := validator.Validate(resource)
	if !result.IsValid() {
		errList = append(errList, result.Errors...)
	}

	ctx := context.Background()
	celValidator := cel.NewValidator(ss, true, celconfig.PerCallLimit)
	errs, _ := celValidator.Validate(ctx, nil, ss, resource.Object, nil, celconfig.RuntimeCELCostBudget)
	for _, e := range errs {
		errList = append(errList, e)
	}

	if len(errList) == 0 {
		return nil
	}

	return errors.CompositeValidationError(errList...)
}
