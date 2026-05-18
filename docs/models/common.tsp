import "@typespec/http";

using TypeSpec.Http;

namespace Api.Models;

/** Standard success response envelope */
model ApiSuccess<T> {
  success: true;
  message: string;
  data: T;
  status_code: int32;
  request_id?: string;
}

/** Standard error response envelope */
model ApiError {
  success: false;
  message: string;
  errors?: unknown;
  status_code: int32;
  request_id?: string;
}

/** 400 Bad Request */
@error
model BadRequest is ApiError;

/** 401 Unauthorized */
@error
model Unauthorized is ApiError;

/** 404 Not Found */
@error
model NotFound is ApiError;

/** 409 Conflict */
@error
model Conflict is ApiError;

/** 422 Unprocessable Entity (validation failure) */
@error
model UnprocessableEntity is ApiError;

/** 500 Internal Server Error */
@error
model InternalServerError is ApiError;
