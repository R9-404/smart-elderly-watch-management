import { accepted, api, type Envelope } from './client'

export interface LatestHealthRecord {
  sampleId: string
  oldId: number
  deviceId: number | null
  bindingVersion: number
  revision: number
  sample: {
    recordId: string
    source: string
    sourceDevice: string
    sourceVersion?: string
    metric: string
    semantics: string
    unit: string
    value: number | null
    start: string
    end: string
    stage?: string
    timeBasis?: string
  }
}

interface LatestHealthReply extends Envelope { oldId: number; items: LatestHealthRecord[] }

export async function readLatestHealth(signal: AbortSignal): Promise<LatestHealthRecord[]> {
  const { data } = await api.post<LatestHealthReply>('api/v1/health/latest', {}, { signal })
  accepted(data)
  return data.items
}
