// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

/*
 * App Name app API
 *
 * API to access and configure the App Name app
 *
 * API version: 1.0.0
 */

package apiserver

// Configuration - Simply a example configuration
type Configuration struct {

	// A id identifying the example configuration
	Id *int64 `json:"id,omitempty"`

	// Configuration data for example
	Config string `json:"config,omitempty"`
}

// AssertConfigurationRequired checks if the required fields are not zero-ed
func AssertConfigurationRequired(obj Configuration) error {
	return nil
}

// AssertConfigurationConstraints checks if the values respects the defined constraints
func AssertConfigurationConstraints(obj Configuration) error {
	return nil
}
