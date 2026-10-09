import axios, { type AxiosRequestConfig, type AxiosResponse } from 'axios'
import { translate } from '@/shared/i18n'
import { useAuthStore } from '@/shared/stores/auth-store'
import {
  AUTH_HEADER,
  AUTH_SCHEME,
  CLIENT_ERROR_CODE,
  CSRF_HEADER,
  DEFAULT_API_BASE_URL,
  DEFAULT_METHOD,
  GATEWAY_STATUSES,
  SAFE_METHODS,
} from './constants'
import { ApiError, type ApiEnvelope, type ApiResult } from './types'

// Latest CSRF token exposed by the server; echoed back on state-changing requests.
let csrfToken: string | null = null

const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? DEFAULT_API_BASE_URL,
  withCredentials: true,
})

http.interceptors.request.use((config) => {
  const { token } = useAuthStore.getState()
  if (token) config.headers.set(AUTH_HEADER, `${AUTH_SCHEME} ${token}`)

  const method = (config.method ?? DEFAULT_METHOD).toUpperCase()
  if (csrfToken && !SAFE_METHODS.has(method)) config.headers.set(CSRF_HEADER, csrfToken)

  return config
})

http.interceptors.response.use(
  (response) => {
    captureCsrfToken(response)
    // `?? {}` covers empty bodies such as 204 No Content.
    const { data, meta } = (response.data ?? {}) as Partial<ApiEnvelope<unknown>>
    // Axios types expect an AxiosResponse here; `requestWithMeta` types the real shape.
    return { data, meta } as unknown as AxiosResponse
  },
  (error: unknown) => {
    if (axios.isAxiosError(error) && error.response) captureCsrfToken(error.response)
    return Promise.reject(toApiError(error))
  },
)

function captureCsrfToken({ headers }: AxiosResponse) {
  // Browsers report response header names in lowercase.
  const token: unknown = headers[CSRF_HEADER.toLowerCase()]
  if (typeof token === 'string' && token) csrfToken = token
}

function toApiError(error: unknown): ApiError {
  if (axios.isAxiosError<Partial<ApiEnvelope<unknown>>>(error) && error.response) {
    const { status, data } = error.response
    return new ApiError({
      status,
      code: data?.meta?.error_code || `HTTP_${status}`,
      message:
        data?.meta?.message || (GATEWAY_STATUSES.has(status) ? translate('The server is unreachable.') : error.message),
      details: data?.data,
    })
  }
  if (axios.isAxiosError(error)) {
    return new ApiError({
      status: 0,
      code: CLIENT_ERROR_CODE.network,
      message: translate('The server is unreachable.'),
      details: error.message,
    })
  }
  return new ApiError({
    status: 0,
    code: CLIENT_ERROR_CODE.unknown,
    message: error instanceof Error ? error.message : translate('Unexpected error.'),
    details: error,
  })
}

/** Sends a request and resolves with `{ data, meta }`, e.g. for paginated lists. */
export function requestWithMeta<T>(config: AxiosRequestConfig): Promise<ApiResult<T>> {
  return http.request<T, ApiResult<T>>(config)
}

/** Sends a request and resolves with the unwrapped `data` only. */
export async function request<T>(config: AxiosRequestConfig): Promise<T> {
  const { data } = await requestWithMeta<T>(config)
  return data
}
