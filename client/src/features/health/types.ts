/** Payload of the server health check (`GET /`). */
export interface Health {
  status: string
  timestamp: string
  version: string
  service: string
}
