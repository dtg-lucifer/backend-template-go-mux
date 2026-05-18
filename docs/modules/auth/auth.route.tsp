import "@typespec/http";
import "./auth.model.tsp";

using TypeSpec.Http;
using Api.Models;

namespace Api.Auth;

@tag("Auth")
@route("/auth")
interface AuthRoutes {
  /** Create a new user account. Returns the created user object. */
  @post
  @route("/register")
  @summary("Register")
  register(@body body: RegisterRequest): {
    @statusCode statusCode: 201;
    @body body: ApiSuccess<RegisterData>;
  } | {
    @statusCode statusCode: 409;
    @body body: Conflict;
  } | {
    @statusCode statusCode: 422;
    @body body: UnprocessableEntity;
  } | {
    @statusCode statusCode: 500;
    @body body: InternalServerError;
  };

  /** Authenticate with email and password. Returns access and refresh tokens. */
  @post
  @route("/login")
  @summary("Login")
  login(@body body: LoginRequest): {
    @statusCode statusCode: 200;
    @body body: ApiSuccess<LoginData>;
  } | {
    @statusCode statusCode: 401;
    @body body: Unauthorized;
  } | {
    @statusCode statusCode: 422;
    @body body: UnprocessableEntity;
  };

  /** Exchange a refresh token for a new access token. */
  @post
  @route("/refresh")
  @summary("Refresh Token")
  refresh(@body body: RefreshRequest): {
    @statusCode statusCode: 200;
    @body body: ApiSuccess<RefreshData>;
  } | {
    @statusCode statusCode: 401;
    @body body: Unauthorized;
  } | {
    @statusCode statusCode: 422;
    @body body: UnprocessableEntity;
  };

  /** Get the currently authenticated user. Requires a valid Bearer token. */
  @get
  @route("/me")
  @summary("Get Current User")
  @useAuth(BearerAuth)
  me(): {
    @statusCode statusCode: 200;
    @body body: ApiSuccess<MeData>;
  } | {
    @statusCode statusCode: 401;
    @body body: Unauthorized;
  } | {
    @statusCode statusCode: 404;
    @body body: NotFound;
  };
}
