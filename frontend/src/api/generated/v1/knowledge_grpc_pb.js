// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var v1_knowledge_pb = require('../v1/knowledge_pb.js');
var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js');
var google_protobuf_field_mask_pb = require('google-protobuf/google/protobuf/field_mask_pb.js');
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');

function serialize_google_protobuf_Empty(arg) {
  if (!(arg instanceof google_protobuf_empty_pb.Empty)) {
    throw new Error('Expected argument of type google.protobuf.Empty');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_google_protobuf_Empty(buffer_arg) {
  return google_protobuf_empty_pb.Empty.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_ConfirmUploadRequest(arg) {
  if (!(arg instanceof v1_knowledge_pb.ConfirmUploadRequest)) {
    throw new Error('Expected argument of type knowledge.v1.ConfirmUploadRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_ConfirmUploadRequest(buffer_arg) {
  return v1_knowledge_pb.ConfirmUploadRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_ContentSource(arg) {
  if (!(arg instanceof v1_knowledge_pb.ContentSource)) {
    throw new Error('Expected argument of type knowledge.v1.ContentSource');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_ContentSource(buffer_arg) {
  return v1_knowledge_pb.ContentSource.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_CreateKnowledgeLinkRequest(arg) {
  if (!(arg instanceof v1_knowledge_pb.CreateKnowledgeLinkRequest)) {
    throw new Error('Expected argument of type knowledge.v1.CreateKnowledgeLinkRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_CreateKnowledgeLinkRequest(buffer_arg) {
  return v1_knowledge_pb.CreateKnowledgeLinkRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_CreateSpaceRequest(arg) {
  if (!(arg instanceof v1_knowledge_pb.CreateSpaceRequest)) {
    throw new Error('Expected argument of type knowledge.v1.CreateSpaceRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_CreateSpaceRequest(buffer_arg) {
  return v1_knowledge_pb.CreateSpaceRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_CreateUploadURLRequest(arg) {
  if (!(arg instanceof v1_knowledge_pb.CreateUploadURLRequest)) {
    throw new Error('Expected argument of type knowledge.v1.CreateUploadURLRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_CreateUploadURLRequest(buffer_arg) {
  return v1_knowledge_pb.CreateUploadURLRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_CreateUploadURLResponse(arg) {
  if (!(arg instanceof v1_knowledge_pb.CreateUploadURLResponse)) {
    throw new Error('Expected argument of type knowledge.v1.CreateUploadURLResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_CreateUploadURLResponse(buffer_arg) {
  return v1_knowledge_pb.CreateUploadURLResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_DeleteKnowledgeLinkRequest(arg) {
  if (!(arg instanceof v1_knowledge_pb.DeleteKnowledgeLinkRequest)) {
    throw new Error('Expected argument of type knowledge.v1.DeleteKnowledgeLinkRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_DeleteKnowledgeLinkRequest(buffer_arg) {
  return v1_knowledge_pb.DeleteKnowledgeLinkRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_DeleteSpaceRequest(arg) {
  if (!(arg instanceof v1_knowledge_pb.DeleteSpaceRequest)) {
    throw new Error('Expected argument of type knowledge.v1.DeleteSpaceRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_DeleteSpaceRequest(buffer_arg) {
  return v1_knowledge_pb.DeleteSpaceRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_EnrichedKnowledgeLink(arg) {
  if (!(arg instanceof v1_knowledge_pb.EnrichedKnowledgeLink)) {
    throw new Error('Expected argument of type knowledge.v1.EnrichedKnowledgeLink');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_EnrichedKnowledgeLink(buffer_arg) {
  return v1_knowledge_pb.EnrichedKnowledgeLink.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_GetBacklinksRequest(arg) {
  if (!(arg instanceof v1_knowledge_pb.GetBacklinksRequest)) {
    throw new Error('Expected argument of type knowledge.v1.GetBacklinksRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_GetBacklinksRequest(buffer_arg) {
  return v1_knowledge_pb.GetBacklinksRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_GetBacklinksResponse(arg) {
  if (!(arg instanceof v1_knowledge_pb.GetBacklinksResponse)) {
    throw new Error('Expected argument of type knowledge.v1.GetBacklinksResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_GetBacklinksResponse(buffer_arg) {
  return v1_knowledge_pb.GetBacklinksResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_GetContentSourceRequest(arg) {
  if (!(arg instanceof v1_knowledge_pb.GetContentSourceRequest)) {
    throw new Error('Expected argument of type knowledge.v1.GetContentSourceRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_GetContentSourceRequest(buffer_arg) {
  return v1_knowledge_pb.GetContentSourceRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_GetKnowledgeLinkRequest(arg) {
  if (!(arg instanceof v1_knowledge_pb.GetKnowledgeLinkRequest)) {
    throw new Error('Expected argument of type knowledge.v1.GetKnowledgeLinkRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_GetKnowledgeLinkRequest(buffer_arg) {
  return v1_knowledge_pb.GetKnowledgeLinkRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_GetSpaceRequest(arg) {
  if (!(arg instanceof v1_knowledge_pb.GetSpaceRequest)) {
    throw new Error('Expected argument of type knowledge.v1.GetSpaceRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_GetSpaceRequest(buffer_arg) {
  return v1_knowledge_pb.GetSpaceRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_HealthStatus(arg) {
  if (!(arg instanceof v1_knowledge_pb.HealthStatus)) {
    throw new Error('Expected argument of type knowledge.v1.HealthStatus');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_HealthStatus(buffer_arg) {
  return v1_knowledge_pb.HealthStatus.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_KnowledgeLink(arg) {
  if (!(arg instanceof v1_knowledge_pb.KnowledgeLink)) {
    throw new Error('Expected argument of type knowledge.v1.KnowledgeLink');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_KnowledgeLink(buffer_arg) {
  return v1_knowledge_pb.KnowledgeLink.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_ListAllSpaceLinksRequest(arg) {
  if (!(arg instanceof v1_knowledge_pb.ListAllSpaceLinksRequest)) {
    throw new Error('Expected argument of type knowledge.v1.ListAllSpaceLinksRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_ListAllSpaceLinksRequest(buffer_arg) {
  return v1_knowledge_pb.ListAllSpaceLinksRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_ListAllSpaceLinksResponse(arg) {
  if (!(arg instanceof v1_knowledge_pb.ListAllSpaceLinksResponse)) {
    throw new Error('Expected argument of type knowledge.v1.ListAllSpaceLinksResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_ListAllSpaceLinksResponse(buffer_arg) {
  return v1_knowledge_pb.ListAllSpaceLinksResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_ListContentSourcesRequest(arg) {
  if (!(arg instanceof v1_knowledge_pb.ListContentSourcesRequest)) {
    throw new Error('Expected argument of type knowledge.v1.ListContentSourcesRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_ListContentSourcesRequest(buffer_arg) {
  return v1_knowledge_pb.ListContentSourcesRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_ListContentSourcesResponse(arg) {
  if (!(arg instanceof v1_knowledge_pb.ListContentSourcesResponse)) {
    throw new Error('Expected argument of type knowledge.v1.ListContentSourcesResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_ListContentSourcesResponse(buffer_arg) {
  return v1_knowledge_pb.ListContentSourcesResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_ListKnowledgeLinksRequest(arg) {
  if (!(arg instanceof v1_knowledge_pb.ListKnowledgeLinksRequest)) {
    throw new Error('Expected argument of type knowledge.v1.ListKnowledgeLinksRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_ListKnowledgeLinksRequest(buffer_arg) {
  return v1_knowledge_pb.ListKnowledgeLinksRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_ListKnowledgeLinksResponse(arg) {
  if (!(arg instanceof v1_knowledge_pb.ListKnowledgeLinksResponse)) {
    throw new Error('Expected argument of type knowledge.v1.ListKnowledgeLinksResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_ListKnowledgeLinksResponse(buffer_arg) {
  return v1_knowledge_pb.ListKnowledgeLinksResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_ListSpacesRequest(arg) {
  if (!(arg instanceof v1_knowledge_pb.ListSpacesRequest)) {
    throw new Error('Expected argument of type knowledge.v1.ListSpacesRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_ListSpacesRequest(buffer_arg) {
  return v1_knowledge_pb.ListSpacesRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_ListSpacesResponse(arg) {
  if (!(arg instanceof v1_knowledge_pb.ListSpacesResponse)) {
    throw new Error('Expected argument of type knowledge.v1.ListSpacesResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_ListSpacesResponse(buffer_arg) {
  return v1_knowledge_pb.ListSpacesResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_Space(arg) {
  if (!(arg instanceof v1_knowledge_pb.Space)) {
    throw new Error('Expected argument of type knowledge.v1.Space');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_Space(buffer_arg) {
  return v1_knowledge_pb.Space.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_UpdateContentSourceStatusRequest(arg) {
  if (!(arg instanceof v1_knowledge_pb.UpdateContentSourceStatusRequest)) {
    throw new Error('Expected argument of type knowledge.v1.UpdateContentSourceStatusRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_UpdateContentSourceStatusRequest(buffer_arg) {
  return v1_knowledge_pb.UpdateContentSourceStatusRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_UpdateKnowledgeLinkRequest(arg) {
  if (!(arg instanceof v1_knowledge_pb.UpdateKnowledgeLinkRequest)) {
    throw new Error('Expected argument of type knowledge.v1.UpdateKnowledgeLinkRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_UpdateKnowledgeLinkRequest(buffer_arg) {
  return v1_knowledge_pb.UpdateKnowledgeLinkRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_knowledge_v1_UpdateSpaceRequest(arg) {
  if (!(arg instanceof v1_knowledge_pb.UpdateSpaceRequest)) {
    throw new Error('Expected argument of type knowledge.v1.UpdateSpaceRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_knowledge_v1_UpdateSpaceRequest(buffer_arg) {
  return v1_knowledge_pb.UpdateSpaceRequest.deserializeBinary(new Uint8Array(buffer_arg));
}


// Knowledge Service provides APIs for managing knowledge spaces, content sources, and knowledge graph links
var KnowledgeServiceService = exports.KnowledgeServiceService = {
  // Space Management
createSpace: {
    path: '/knowledge.v1.KnowledgeService/CreateSpace',
    requestStream: false,
    responseStream: false,
    requestType: v1_knowledge_pb.CreateSpaceRequest,
    responseType: v1_knowledge_pb.Space,
    requestSerialize: serialize_knowledge_v1_CreateSpaceRequest,
    requestDeserialize: deserialize_knowledge_v1_CreateSpaceRequest,
    responseSerialize: serialize_knowledge_v1_Space,
    responseDeserialize: deserialize_knowledge_v1_Space,
  },
  getSpace: {
    path: '/knowledge.v1.KnowledgeService/GetSpace',
    requestStream: false,
    responseStream: false,
    requestType: v1_knowledge_pb.GetSpaceRequest,
    responseType: v1_knowledge_pb.Space,
    requestSerialize: serialize_knowledge_v1_GetSpaceRequest,
    requestDeserialize: deserialize_knowledge_v1_GetSpaceRequest,
    responseSerialize: serialize_knowledge_v1_Space,
    responseDeserialize: deserialize_knowledge_v1_Space,
  },
  listSpaces: {
    path: '/knowledge.v1.KnowledgeService/ListSpaces',
    requestStream: false,
    responseStream: false,
    requestType: v1_knowledge_pb.ListSpacesRequest,
    responseType: v1_knowledge_pb.ListSpacesResponse,
    requestSerialize: serialize_knowledge_v1_ListSpacesRequest,
    requestDeserialize: deserialize_knowledge_v1_ListSpacesRequest,
    responseSerialize: serialize_knowledge_v1_ListSpacesResponse,
    responseDeserialize: deserialize_knowledge_v1_ListSpacesResponse,
  },
  updateSpace: {
    path: '/knowledge.v1.KnowledgeService/UpdateSpace',
    requestStream: false,
    responseStream: false,
    requestType: v1_knowledge_pb.UpdateSpaceRequest,
    responseType: v1_knowledge_pb.Space,
    requestSerialize: serialize_knowledge_v1_UpdateSpaceRequest,
    requestDeserialize: deserialize_knowledge_v1_UpdateSpaceRequest,
    responseSerialize: serialize_knowledge_v1_Space,
    responseDeserialize: deserialize_knowledge_v1_Space,
  },
  deleteSpace: {
    path: '/knowledge.v1.KnowledgeService/DeleteSpace',
    requestStream: false,
    responseStream: false,
    requestType: v1_knowledge_pb.DeleteSpaceRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_knowledge_v1_DeleteSpaceRequest,
    requestDeserialize: deserialize_knowledge_v1_DeleteSpaceRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  // Content Source Management
createUploadURL: {
    path: '/knowledge.v1.KnowledgeService/CreateUploadURL',
    requestStream: false,
    responseStream: false,
    requestType: v1_knowledge_pb.CreateUploadURLRequest,
    responseType: v1_knowledge_pb.CreateUploadURLResponse,
    requestSerialize: serialize_knowledge_v1_CreateUploadURLRequest,
    requestDeserialize: deserialize_knowledge_v1_CreateUploadURLRequest,
    responseSerialize: serialize_knowledge_v1_CreateUploadURLResponse,
    responseDeserialize: deserialize_knowledge_v1_CreateUploadURLResponse,
  },
  confirmUpload: {
    path: '/knowledge.v1.KnowledgeService/ConfirmUpload',
    requestStream: false,
    responseStream: false,
    requestType: v1_knowledge_pb.ConfirmUploadRequest,
    responseType: v1_knowledge_pb.ContentSource,
    requestSerialize: serialize_knowledge_v1_ConfirmUploadRequest,
    requestDeserialize: deserialize_knowledge_v1_ConfirmUploadRequest,
    responseSerialize: serialize_knowledge_v1_ContentSource,
    responseDeserialize: deserialize_knowledge_v1_ContentSource,
  },
  getContentSource: {
    path: '/knowledge.v1.KnowledgeService/GetContentSource',
    requestStream: false,
    responseStream: false,
    requestType: v1_knowledge_pb.GetContentSourceRequest,
    responseType: v1_knowledge_pb.ContentSource,
    requestSerialize: serialize_knowledge_v1_GetContentSourceRequest,
    requestDeserialize: deserialize_knowledge_v1_GetContentSourceRequest,
    responseSerialize: serialize_knowledge_v1_ContentSource,
    responseDeserialize: deserialize_knowledge_v1_ContentSource,
  },
  listContentSources: {
    path: '/knowledge.v1.KnowledgeService/ListContentSources',
    requestStream: false,
    responseStream: false,
    requestType: v1_knowledge_pb.ListContentSourcesRequest,
    responseType: v1_knowledge_pb.ListContentSourcesResponse,
    requestSerialize: serialize_knowledge_v1_ListContentSourcesRequest,
    requestDeserialize: deserialize_knowledge_v1_ListContentSourcesRequest,
    responseSerialize: serialize_knowledge_v1_ListContentSourcesResponse,
    responseDeserialize: deserialize_knowledge_v1_ListContentSourcesResponse,
  },
  updateContentSourceStatus: {
    path: '/knowledge.v1.KnowledgeService/UpdateContentSourceStatus',
    requestStream: false,
    responseStream: false,
    requestType: v1_knowledge_pb.UpdateContentSourceStatusRequest,
    responseType: v1_knowledge_pb.ContentSource,
    requestSerialize: serialize_knowledge_v1_UpdateContentSourceStatusRequest,
    requestDeserialize: deserialize_knowledge_v1_UpdateContentSourceStatusRequest,
    responseSerialize: serialize_knowledge_v1_ContentSource,
    responseDeserialize: deserialize_knowledge_v1_ContentSource,
  },
  // Knowledge Graph Management
createKnowledgeLink: {
    path: '/knowledge.v1.KnowledgeService/CreateKnowledgeLink',
    requestStream: false,
    responseStream: false,
    requestType: v1_knowledge_pb.CreateKnowledgeLinkRequest,
    responseType: v1_knowledge_pb.KnowledgeLink,
    requestSerialize: serialize_knowledge_v1_CreateKnowledgeLinkRequest,
    requestDeserialize: deserialize_knowledge_v1_CreateKnowledgeLinkRequest,
    responseSerialize: serialize_knowledge_v1_KnowledgeLink,
    responseDeserialize: deserialize_knowledge_v1_KnowledgeLink,
  },
  // Single enriched link by id
getKnowledgeLink: {
    path: '/knowledge.v1.KnowledgeService/GetKnowledgeLink',
    requestStream: false,
    responseStream: false,
    requestType: v1_knowledge_pb.GetKnowledgeLinkRequest,
    responseType: v1_knowledge_pb.EnrichedKnowledgeLink,
    requestSerialize: serialize_knowledge_v1_GetKnowledgeLinkRequest,
    requestDeserialize: deserialize_knowledge_v1_GetKnowledgeLinkRequest,
    responseSerialize: serialize_knowledge_v1_EnrichedKnowledgeLink,
    responseDeserialize: deserialize_knowledge_v1_EnrichedKnowledgeLink,
  },
  // Focused enriched list by content
listKnowledgeLinks: {
    path: '/knowledge.v1.KnowledgeService/ListKnowledgeLinks',
    requestStream: false,
    responseStream: false,
    requestType: v1_knowledge_pb.ListKnowledgeLinksRequest,
    responseType: v1_knowledge_pb.ListKnowledgeLinksResponse,
    requestSerialize: serialize_knowledge_v1_ListKnowledgeLinksRequest,
    requestDeserialize: deserialize_knowledge_v1_ListKnowledgeLinksRequest,
    responseSerialize: serialize_knowledge_v1_ListKnowledgeLinksResponse,
    responseDeserialize: deserialize_knowledge_v1_ListKnowledgeLinksResponse,
  },
  // Bulk simple list by space
listAllSpaceLinks: {
    path: '/knowledge.v1.KnowledgeService/ListAllSpaceLinks',
    requestStream: false,
    responseStream: false,
    requestType: v1_knowledge_pb.ListAllSpaceLinksRequest,
    responseType: v1_knowledge_pb.ListAllSpaceLinksResponse,
    requestSerialize: serialize_knowledge_v1_ListAllSpaceLinksRequest,
    requestDeserialize: deserialize_knowledge_v1_ListAllSpaceLinksRequest,
    responseSerialize: serialize_knowledge_v1_ListAllSpaceLinksResponse,
    responseDeserialize: deserialize_knowledge_v1_ListAllSpaceLinksResponse,
  },
  updateKnowledgeLink: {
    path: '/knowledge.v1.KnowledgeService/UpdateKnowledgeLink',
    requestStream: false,
    responseStream: false,
    requestType: v1_knowledge_pb.UpdateKnowledgeLinkRequest,
    responseType: v1_knowledge_pb.KnowledgeLink,
    requestSerialize: serialize_knowledge_v1_UpdateKnowledgeLinkRequest,
    requestDeserialize: deserialize_knowledge_v1_UpdateKnowledgeLinkRequest,
    responseSerialize: serialize_knowledge_v1_KnowledgeLink,
    responseDeserialize: deserialize_knowledge_v1_KnowledgeLink,
  },
  deleteKnowledgeLink: {
    path: '/knowledge.v1.KnowledgeService/DeleteKnowledgeLink',
    requestStream: false,
    responseStream: false,
    requestType: v1_knowledge_pb.DeleteKnowledgeLinkRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_knowledge_v1_DeleteKnowledgeLinkRequest,
    requestDeserialize: deserialize_knowledge_v1_DeleteKnowledgeLinkRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  getBacklinks: {
    path: '/knowledge.v1.KnowledgeService/GetBacklinks',
    requestStream: false,
    responseStream: false,
    requestType: v1_knowledge_pb.GetBacklinksRequest,
    responseType: v1_knowledge_pb.GetBacklinksResponse,
    requestSerialize: serialize_knowledge_v1_GetBacklinksRequest,
    requestDeserialize: deserialize_knowledge_v1_GetBacklinksRequest,
    responseSerialize: serialize_knowledge_v1_GetBacklinksResponse,
    responseDeserialize: deserialize_knowledge_v1_GetBacklinksResponse,
  },
  // Utilities
healthz: {
    path: '/knowledge.v1.KnowledgeService/Healthz',
    requestStream: false,
    responseStream: false,
    requestType: google_protobuf_empty_pb.Empty,
    responseType: v1_knowledge_pb.HealthStatus,
    requestSerialize: serialize_google_protobuf_Empty,
    requestDeserialize: deserialize_google_protobuf_Empty,
    responseSerialize: serialize_knowledge_v1_HealthStatus,
    responseDeserialize: deserialize_knowledge_v1_HealthStatus,
  },
};

exports.KnowledgeServiceClient = grpc.makeGenericClientConstructor(KnowledgeServiceService, 'KnowledgeService');
