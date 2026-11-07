import type { Envelope } from './client'
import type { PositionContext } from '../utils/positionContext'

export type RecordId = number | string
export type AlertState = number | string | null
export interface ActivityRecord {
  oldId: RecordId
  name: string | null
  step: number | null
  sleep: number | string | null
  sleepAlertState: AlertState
  createTime: string | null
}
export interface VitalRecord {
  oldId: RecordId
  name: string | null
  preState: AlertState
  rateState: AlertState
  createTime: string | null
}
export interface LocationRecord extends PositionContext {
  oldId: RecordId
  name: string | null
  lng: string | number | null
  lat: string | number | null
  status: AlertState
  beTime: string | null
}
export interface FamilyProfile {
  familyId?: RecordId
  familyName?: string | null
  familySex?: string | null
  familyAddress?: string | null
  familyTouPic?: string | null
}
export interface ResidentRecord extends Record<string, unknown> {
  oldId: RecordId
  name: string | null
  sex: string | null
  bothdate: string | null
  phone: string | null
  room: string | null
  beTime: string | null
  customerId: RecordId | null
  customerName: string | null
  deviceId: RecordId | null
  illness: string | null
  warn: string | null
  foodLike: string | null
  foodHate: string | null
  drag: string | null
}
export interface GuardianOption {
  customerId: RecordId
  customerName: string
}
export interface CareTask {
  id: number
  oldId: number
  deviceId: number
  kind: string
  state: string
  severity: string
  oldPersonName: string | null
  notificationState: string
  createdAt: string
  dueAt: string
  verifiedAt: string | null
  acknowledgedBy: string | null
  resolvedBy: string | null
}
export interface CareEvent {
  action: string
  createdAt: string
  actor: string | null
  note: string | null
}
export interface CareTaskReply extends Envelope {
  items: CareTask[]
  total: number
  duty: string
  actorId: number
}
export interface CareHistoryReply extends Envelope { events: CareEvent[]; total: number }
export interface CareDiagnostics extends Envelope {
  primaryEngine: string
  workerHealthy: boolean
  notificationConfigured: boolean
  receiptConfigured: boolean
  pendingNotifications: number
  failedNotifications: number
  staleSending: number
  overdueIncidents: number
  analytics: { configured: boolean; ready: boolean; pending: number | null; lastSyncedAt?: string | null }
}
export interface CareStaff { id: number; name: string | null; duty: string }
export interface CareAccess { actorId: number; resource: string; at: string }
