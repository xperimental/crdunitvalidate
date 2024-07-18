package crdunitvalidate

import (
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func LoadYAML(bytes []byte) (*unstructured.Unstructured, error) {
	jsonBytes, err := yaml.YAMLToJSON(bytes)
	if err != nil {
		return nil, err
	}

	return LoadJSON(jsonBytes)
}

func LoadJSON(bytes []byte) (*unstructured.Unstructured, error) {
	obj, _, err := unstructured.UnstructuredJSONScheme.Decode(bytes, nil, nil)
	if err != nil {
		return nil, err
	}

	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("cannot convert to unstructured: %t", obj)
	}

	return u, nil
}
