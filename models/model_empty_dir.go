/*
 * Delegate Runner
 *
 * This is the API for Delegate Runner
 *
 * API version: 1.0.11
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package models

type EmptyDir struct {

	VolumeId string `json:"volumeId,omitempty"`

	Name string `json:"name,omitempty"`

	Medium string `json:"medium,omitempty"`

	SizeLimit int32 `json:"sizeLimit,omitempty"`

	Labels map[string]string `json:"labels,omitempty"`
}
