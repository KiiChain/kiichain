package utils_test

// Hiding this because we still have references to the Oracle module on Cosmos-SDK
// TODO: FIX ME AND ENABLE ME
// func TestAllResourcesInTree(t *testing.T) {
// 	storeKeyToResourceMap := aclutils.StoreKeyToResourceTypePrefixMap
// 	resourceTree := sdkacltypes.ResourceTree

// 	storeKeyAllResourceTypes := make(map[sdkacltypes.ResourceType]bool)
// 	for _, resourceTypeToPrefix := range storeKeyToResourceMap {
// 		for resourceType := range resourceTypeToPrefix {
// 			storeKeyAllResourceTypes[resourceType] = true
// 		}
// 	}

// 	for resourceType := range resourceTree {
// 		if _, ok := storeKeyAllResourceTypes[resourceType]; !ok {
// 			panic(fmt.Sprintf("Missing resourceType=%s in the storekey to resource type prefix mapping", resourceType))
// 		}
// 	}
// }

// Hiding this because we still have references to the Oracle module on Cosmos-SDK
// TODO: FIX ME AND ENABLE ME
// func TestAllResourcesInStoreKeyMap(t *testing.T) {
// 	resourceToStoreKeyMap := aclutils.ResourceTypeToStoreKeyMap
// 	resourceTree := sdkacltypes.ResourceTree

// 	storeKeyAllResourceTypes := make(map[sdkacltypes.ResourceType]bool)
// 	for resourceType := range resourceToStoreKeyMap {
// 		storeKeyAllResourceTypes[resourceType] = true
// 	}

// 	for resourceType := range resourceTree {
// 		// omit ANY, KV, and MEM
// 		if resourceType == sdkacltypes.ResourceType_ANY || resourceType == sdkacltypes.ResourceType_KV || resourceType == sdkacltypes.ResourceType_Mem {
// 			continue
// 		}
// 		if _, ok := storeKeyAllResourceTypes[resourceType]; !ok {
// 			panic(fmt.Sprintf("Missing resourceType=%s in the storekey to resource type prefix mapping", resourceType))
// 		}
// 	}

// }
