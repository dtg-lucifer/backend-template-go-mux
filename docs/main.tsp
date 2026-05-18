import "@typespec/http";
import "@typespec/openapi";
import "@typespec/openapi3";
import "./models/common.tsp";
import "./models/auth.tsp";
import "./models/health.tsp";
import "./routes/auth.tsp";
import "./routes/health.tsp";

using TypeSpec.Http;
using TypeSpec.OpenAPI;

@service(#{ title: "Go Mux Backend API" })
@info(#{
  version: "1.0.0",
  contact: #{
    name: "API Support",
    email: "support@example.com",
  },
  license: #{
    name: "MIT",
    url: "https://opensource.org/licenses/MIT",
  },
})
@server("http://localhost:8080/api/v1", "Local development server")
namespace Api;
