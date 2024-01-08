/*
 * Service Manager
 *
 * Service Manager provides REST APIs that are responsible for the creation and consumption of service instances in any connected runtime environment.   Use the Service Manager APIs to perform various operations related to your platforms, service brokers, service instances, and service bindings.  Get service plans and service offerings associated with your environment.    #### Platforms   Platforms are OSBAPI-enabled software systems on which applications and services are hosted.   With the Service Manager, you can now register your platform and enable it to consume the SAP BTP services from your native environment.   This registration results in a returned set of credentials that are needed to deploy the Service Manager agent.     #### Service Brokers   Service brokers act as brokers between the Service Manager and a platform’s marketplace to advertise catalogues of service offerings and service plans.  They also receive and process the requests from the marketplace to provision, bind, unbind, and deprovision these offerings and plans.    #### Service Instances   Service instances are instantiations of service plans that make the functionality of those service plans available for consumption.    #### Service Bindings   Service bindings provide access details to existing service instances.  The access details are part of the service bindings' ‘credentials’ property, and typically include access URLs and credentials.    #### Service Plans   Service plans represent sets of capabilities provided by a service offering.  For example, database service offerings provide different plans for different database versions or sizes, while the Service Manager plans offer different data access levels.    #### Service Offerings   Service offerings are advertisements of the services that are supported by a service broker.  For example, software that you can consume in the subaccount.  Service offerings are related to one or more service plans.
 *
 * API version: 1.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 * ATTENTION MANUAL CHANGES FOR "Errors" parameter - for details see NGPBUG-371222
 */
package servicemanager

import (
	"encoding/json"
	"time"
)

type OperationResponseObject struct {
	// The ID of the operation.
	Id string `json:"id,omitempty"`
	// Whether the resource is ready.
	Ready bool `json:"ready,omitempty"`
	// The type of the operation.<br><br> Possible values:
	Type_ string `json:"type,omitempty"`
	// Valid values are: in progress, succeeded, and failed. <br> While the state is \"in progress\", the platform should continue polling. <br/> The responses: \"state\": \"succeeded\" or \"state\": \"failed\" must cause the platform to stop polling.
	State string `json:"state,omitempty"`
	// Details about the operation for customer-facing UI.
	Description string `json:"description,omitempty"`
	// The ID of the resource. <br> Exists if: \"state\": \"succeeded\", and also for PATCH and DELETE requests
	ResourceId          string               `json:"resource_id,omitempty"`
	TransitiveResources []TransitiveResource `json:"transitive_resources,omitempty"`
	// The type of the resource (e.g. service_brokers, service_instances).
	ResourceType string `json:"resource_type,omitempty"`
	// The ID of the platform associated with the operation.
	PlatformId string `json:"platform_id,omitempty"`
	// The correlation ID received from the request related to this operation.
	CorrelationId string `json:"correlation_id,omitempty"`
	// Whether the operation has reached a checkpoint and can be executed again.
	Reschedule bool `json:"reschedule,omitempty"`
	// The time the resource is scheduled for deletion. <br/>In ISO 8601 format:</br> YYYY-MM-DDThh:mm:ssTZD
	DeletionScheduled time.Time `json:"deletion_scheduled,omitempty"`
	// The time the resource was created. <br/>In ISO 8601 format.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// The last time the resource was updated. <br/>In ISO 8601 format. <br/>Recommended field if \"state\": \"succeeded\" or \"state\": \"failed\".
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// The list of the errors if the operation has failed.
	// Manually changed from type []ModelError to generic json.RawMessage due to NGPBUG-371222
	Errors json.RawMessage      `json:"errors,omitempty"`
	Labels *map[string][]string `json:"labels,omitempty"`
}
