# IoArgoprojWorkflowV1alpha1Mutex

Mutex holds Mutex configuration

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **str** | name of the mutex | [optional] 
**namespace** | **str** | Namespace is the namespace of the mutex, default: [namespace of workflow] | [optional] 
**selectors** | [**[IoArgoprojWorkflowV1alpha1SyncSelector]**](IoArgoprojWorkflowV1alpha1SyncSelector.md) | Selectors is a list of references to dynamic values (like parameters, labels, annotations) that can be added to mutex key to make concurrency more customizable | [optional] 
**any string name** | **bool, date, datetime, dict, float, int, list, str, none_type** | any string name can be used but the value must be the correct type | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


