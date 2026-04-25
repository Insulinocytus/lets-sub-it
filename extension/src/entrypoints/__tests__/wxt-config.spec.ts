import { describe, expect, it } from 'vitest'
import config from '../../../wxt.config'

describe('wxt manifest host permissions', () => {
  it('allows the backend host used by extension code', () => {
    expect(config.manifest?.host_permissions).toContain('http://127.0.0.1:8080/*')
  })
})
