// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var v1_user_pb = require('../v1/user_pb.js');
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');
var google_protobuf_struct_pb = require('google-protobuf/google/protobuf/struct_pb.js');

function serialize_user_v1_ActivateUserRequest(arg) {
  if (!(arg instanceof v1_user_pb.ActivateUserRequest)) {
    throw new Error('Expected argument of type user.v1.ActivateUserRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_user_v1_ActivateUserRequest(buffer_arg) {
  return v1_user_pb.ActivateUserRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_user_v1_ActivateUserResponse(arg) {
  if (!(arg instanceof v1_user_pb.ActivateUserResponse)) {
    throw new Error('Expected argument of type user.v1.ActivateUserResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_user_v1_ActivateUserResponse(buffer_arg) {
  return v1_user_pb.ActivateUserResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_user_v1_CheckUserStatusRequest(arg) {
  if (!(arg instanceof v1_user_pb.CheckUserStatusRequest)) {
    throw new Error('Expected argument of type user.v1.CheckUserStatusRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_user_v1_CheckUserStatusRequest(buffer_arg) {
  return v1_user_pb.CheckUserStatusRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_user_v1_CheckUserStatusResponse(arg) {
  if (!(arg instanceof v1_user_pb.CheckUserStatusResponse)) {
    throw new Error('Expected argument of type user.v1.CheckUserStatusResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_user_v1_CheckUserStatusResponse(buffer_arg) {
  return v1_user_pb.CheckUserStatusResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_user_v1_CreateUserRequest(arg) {
  if (!(arg instanceof v1_user_pb.CreateUserRequest)) {
    throw new Error('Expected argument of type user.v1.CreateUserRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_user_v1_CreateUserRequest(buffer_arg) {
  return v1_user_pb.CreateUserRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_user_v1_CreateUserResponse(arg) {
  if (!(arg instanceof v1_user_pb.CreateUserResponse)) {
    throw new Error('Expected argument of type user.v1.CreateUserResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_user_v1_CreateUserResponse(buffer_arg) {
  return v1_user_pb.CreateUserResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_user_v1_DeleteUserRequest(arg) {
  if (!(arg instanceof v1_user_pb.DeleteUserRequest)) {
    throw new Error('Expected argument of type user.v1.DeleteUserRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_user_v1_DeleteUserRequest(buffer_arg) {
  return v1_user_pb.DeleteUserRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_user_v1_DeleteUserResponse(arg) {
  if (!(arg instanceof v1_user_pb.DeleteUserResponse)) {
    throw new Error('Expected argument of type user.v1.DeleteUserResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_user_v1_DeleteUserResponse(buffer_arg) {
  return v1_user_pb.DeleteUserResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_user_v1_GetUserRequest(arg) {
  if (!(arg instanceof v1_user_pb.GetUserRequest)) {
    throw new Error('Expected argument of type user.v1.GetUserRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_user_v1_GetUserRequest(buffer_arg) {
  return v1_user_pb.GetUserRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_user_v1_GetUserResponse(arg) {
  if (!(arg instanceof v1_user_pb.GetUserResponse)) {
    throw new Error('Expected argument of type user.v1.GetUserResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_user_v1_GetUserResponse(buffer_arg) {
  return v1_user_pb.GetUserResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_user_v1_UpdateUserPreferencesRequest(arg) {
  if (!(arg instanceof v1_user_pb.UpdateUserPreferencesRequest)) {
    throw new Error('Expected argument of type user.v1.UpdateUserPreferencesRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_user_v1_UpdateUserPreferencesRequest(buffer_arg) {
  return v1_user_pb.UpdateUserPreferencesRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_user_v1_UpdateUserPreferencesResponse(arg) {
  if (!(arg instanceof v1_user_pb.UpdateUserPreferencesResponse)) {
    throw new Error('Expected argument of type user.v1.UpdateUserPreferencesResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_user_v1_UpdateUserPreferencesResponse(buffer_arg) {
  return v1_user_pb.UpdateUserPreferencesResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_user_v1_UpdateUserRequest(arg) {
  if (!(arg instanceof v1_user_pb.UpdateUserRequest)) {
    throw new Error('Expected argument of type user.v1.UpdateUserRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_user_v1_UpdateUserRequest(buffer_arg) {
  return v1_user_pb.UpdateUserRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_user_v1_UpdateUserResponse(arg) {
  if (!(arg instanceof v1_user_pb.UpdateUserResponse)) {
    throw new Error('Expected argument of type user.v1.UpdateUserResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_user_v1_UpdateUserResponse(buffer_arg) {
  return v1_user_pb.UpdateUserResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


// UserService manages user accounts and preferences.
var UserServiceService = exports.UserServiceService = {
  // Creates a new user with pending status (called by webhook).
createUser: {
    path: '/user.v1.UserService/CreateUser',
    requestStream: false,
    responseStream: false,
    requestType: v1_user_pb.CreateUserRequest,
    responseType: v1_user_pb.CreateUserResponse,
    requestSerialize: serialize_user_v1_CreateUserRequest,
    requestDeserialize: deserialize_user_v1_CreateUserRequest,
    responseSerialize: serialize_user_v1_CreateUserResponse,
    responseDeserialize: deserialize_user_v1_CreateUserResponse,
  },
  // Retrieves a user's profile.
getUser: {
    path: '/user.v1.UserService/GetUser',
    requestStream: false,
    responseStream: false,
    requestType: v1_user_pb.GetUserRequest,
    responseType: v1_user_pb.GetUserResponse,
    requestSerialize: serialize_user_v1_GetUserRequest,
    requestDeserialize: deserialize_user_v1_GetUserRequest,
    responseSerialize: serialize_user_v1_GetUserResponse,
    responseDeserialize: deserialize_user_v1_GetUserResponse,
  },
  // Updates a user's profile information.
updateUser: {
    path: '/user.v1.UserService/UpdateUser',
    requestStream: false,
    responseStream: false,
    requestType: v1_user_pb.UpdateUserRequest,
    responseType: v1_user_pb.UpdateUserResponse,
    requestSerialize: serialize_user_v1_UpdateUserRequest,
    requestDeserialize: deserialize_user_v1_UpdateUserRequest,
    responseSerialize: serialize_user_v1_UpdateUserResponse,
    responseDeserialize: deserialize_user_v1_UpdateUserResponse,
  },
  // Deletes a user (soft delete).
deleteUser: {
    path: '/user.v1.UserService/DeleteUser',
    requestStream: false,
    responseStream: false,
    requestType: v1_user_pb.DeleteUserRequest,
    responseType: v1_user_pb.DeleteUserResponse,
    requestSerialize: serialize_user_v1_DeleteUserRequest,
    requestDeserialize: deserialize_user_v1_DeleteUserRequest,
    responseSerialize: serialize_user_v1_DeleteUserResponse,
    responseDeserialize: deserialize_user_v1_DeleteUserResponse,
  },
  // Updates a user's preferences.
updateUserPreferences: {
    path: '/user.v1.UserService/UpdateUserPreferences',
    requestStream: false,
    responseStream: false,
    requestType: v1_user_pb.UpdateUserPreferencesRequest,
    responseType: v1_user_pb.UpdateUserPreferencesResponse,
    requestSerialize: serialize_user_v1_UpdateUserPreferencesRequest,
    requestDeserialize: deserialize_user_v1_UpdateUserPreferencesRequest,
    responseSerialize: serialize_user_v1_UpdateUserPreferencesResponse,
    responseDeserialize: deserialize_user_v1_UpdateUserPreferencesResponse,
  },
  // Activates a user profile after initial setup (sets status to active).
activateUser: {
    path: '/user.v1.UserService/ActivateUser',
    requestStream: false,
    responseStream: false,
    requestType: v1_user_pb.ActivateUserRequest,
    responseType: v1_user_pb.ActivateUserResponse,
    requestSerialize: serialize_user_v1_ActivateUserRequest,
    requestDeserialize: deserialize_user_v1_ActivateUserRequest,
    responseSerialize: serialize_user_v1_ActivateUserResponse,
    responseDeserialize: deserialize_user_v1_ActivateUserResponse,
  },
  // Checks user status and profile completion.
checkUserStatus: {
    path: '/user.v1.UserService/CheckUserStatus',
    requestStream: false,
    responseStream: false,
    requestType: v1_user_pb.CheckUserStatusRequest,
    responseType: v1_user_pb.CheckUserStatusResponse,
    requestSerialize: serialize_user_v1_CheckUserStatusRequest,
    requestDeserialize: deserialize_user_v1_CheckUserStatusRequest,
    responseSerialize: serialize_user_v1_CheckUserStatusResponse,
    responseDeserialize: deserialize_user_v1_CheckUserStatusResponse,
  },
};

exports.UserServiceClient = grpc.makeGenericClientConstructor(UserServiceService, 'UserService');
