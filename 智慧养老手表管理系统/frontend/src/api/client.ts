import axios from 'axios'
import { isSuccessEnvelope } from './httpContract'

export interface Envelope { returnCode: '1111' | '0000'; code?: string; message?: string; reference?: string }
export interface PostClient { post<T>(path: string, body: unknown, options?: { signal?: AbortSignal }): Promise<{ data: T }> }

const API_ORIGIN = 'https://suining.invalid'

export function apiPath(path: string): string {
  if (typeof path !== 'string' || path.trim() === '') throw new Error('Invalid application API path')
  const candidate = path.startsWith('/') ? path : `/${path}`
  if (/[?#\\]/.test(candidate) || /%(?:2f|5c)/i.test(candidate)) throw new Error('Invalid application API path')
  const resolved = new URL(candidate, API_ORIGIN)
  if (resolved.origin !== API_ORIGIN || resolved.pathname !== candidate || !resolved.pathname.startsWith('/api/v1/')) {
    throw new Error('Invalid application API path')
  }
  return resolved.pathname
}

export const api: PostClient = {
  post: <T>(path: string, body: unknown, options?: { signal?: AbortSignal }) => axios.post<T>(apiPath(path), body, { withCredentials: true, signal: options?.signal })
}
export function accepted(data: Envelope): void {
  if (!isSuccessEnvelope(data)) throw new Error(data?.message || '请求被拒绝', { cause: data })
}
