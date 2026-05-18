import "@typespec/http";
import "./common.tsp";

using TypeSpec.Http;
using Api.Models;

namespace Api.Models;

/** Safe user object returned in responses (no password field) */
model SafeUser {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  verified: boolean;
  created_at: string;
}

// ── Request bodies ────────────────────────────────────────────────────────────

model RegisterRequest {
  first_name: string;
  last_name: string;

  @format("email")
  email: string;

  /** Minimum 8 characters */
  password: string;
}

model LoginRequest {
  @format("email")
  email: string;

  password: string;
}

model RefreshRequest {
  refresh_token: string;
}

// ── Response data shapes ──────────────────────────────────────────────────────

model RegisterData {
  user: SafeUser;
}

model LoginData {
  user: SafeUser;
  access_token: string;
  refresh_token: string;
}

model MeData {
  user: SafeUser;
}

model RefreshData {
  access_token: string;
}
