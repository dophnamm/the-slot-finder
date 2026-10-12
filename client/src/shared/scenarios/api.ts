import { SCENARIO_ENDPOINTS } from '@/resources/endpoints'
import { request } from '@/shared/api/http'
import type { Scenario, ScenarioDetail, ScenarioInput } from './types'

export function listScenarios() {
  return request<Scenario[]>({ url: SCENARIO_ENDPOINTS.list })
}

export function getScenario(id: string) {
  return request<ScenarioDetail>({ url: SCENARIO_ENDPOINTS.detail(id) })
}

export function createScenario(input: ScenarioInput) {
  return request<ScenarioDetail>({ url: SCENARIO_ENDPOINTS.list, method: 'POST', data: input })
}

export function rerunScenario(id: string) {
  return request<ScenarioDetail>({ url: SCENARIO_ENDPOINTS.rerun(id), method: 'POST' })
}

export function deleteScenario(id: string) {
  return request<null>({ url: SCENARIO_ENDPOINTS.detail(id), method: 'DELETE' })
}
