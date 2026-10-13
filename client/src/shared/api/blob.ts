/**
 * A request made with `responseType: 'blob'` receives even its error envelope as a Blob. This
 * reads it back as JSON so the error keeps the server's code and message; a Blob that is not
 * JSON gives undefined, and any other data is returned unchanged.
 */
export async function readBlobEnvelope(data: unknown): Promise<unknown> {
  if (!(data instanceof Blob)) return data
  try {
    return JSON.parse(await data.text()) as unknown
  } catch {
    return undefined
  }
}
