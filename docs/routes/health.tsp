import "@typespec/http";
import "../models/health.tsp";
import "../models/common.tsp";

using TypeSpec.Http;
using Api.Models;

namespace Api.Health;

@tag("Health")
@route("/health")
interface HealthRoutes {
  /** Check the health of the API and its dependencies (database, redis, memory). */
  @get
  @summary("Health Check")
  check(): {
    @statusCode statusCode: 200;
    @body body: HealthData;
  } | {
    @statusCode statusCode: 503;
    @body body: InternalServerError;
  };
}
