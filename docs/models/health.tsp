import "@typespec/http";

using TypeSpec.Http;

namespace Api.Models;

/** Database service status */
model DatabaseStatus {
  status: string;
  open_conns: int32;
  idle_conns: int32;
}

/** Redis service status */
model RedisStatus {
  status: string;
}

/** Memory statistics */
model MemoryStats {
  alloc_mb: float64;
  total_alloc_mb: float64;
  sys_mb: float64;
  num_gc: int32;
}

/** Health check services */
model HealthServices {
  database: DatabaseStatus;
  redis: RedisStatus;
}

/** Health check response data */
model HealthData {
  success: boolean;
  status: string;
  uptime: string;
  services: HealthServices;
  memory: MemoryStats;
  request_id?: string;
}
