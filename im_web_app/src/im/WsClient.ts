import type { Msg } from '@/api/types'

type PushHandler = (msg: Msg) => void
type StatusHandler = (connected: boolean) => void

interface Frame {
  cmd: string
  req_id?: string
  data?: unknown
}

export class WsClient {
  private ws: WebSocket | null = null
  private token = ''
  private heartbeatTimer: number | null = null
  private reconnectTimer: number | null = null
  private closedByUser = false
  private onPush: PushHandler | null = null
  private onStatus: StatusHandler | null = null
  private reqSeq = 1

  setHandlers(onPush: PushHandler, onStatus: StatusHandler): void {
    this.onPush = onPush
    this.onStatus = onStatus
  }

  connect(accessToken: string): void {
    this.token = accessToken
    this.closedByUser = false
    this.open()
  }

  /** 下次重连使用新 token。当前连接继续用旧 token，直到断开。 */
  replaceToken(accessToken: string): void {
    this.token = accessToken
  }

  disconnect(): void {
    this.closedByUser = true
    this.clearTimers()
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.onStatus?.(false)
  }

  private open(): void {
    this.clearTimers()
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    const url = `${proto}://${location.host}/im/ws?access_token=${encodeURIComponent(this.token)}`
    const ws = new WebSocket(url)
    this.ws = ws
    ws.onopen = () => {
      this.onStatus?.(true)
      this.startHeartbeat()
    }
    ws.onmessage = (ev) => {
      try {
        const frame = JSON.parse(String(ev.data)) as Frame
        if (frame.cmd === 'msg.push' && frame.data) {
          this.onPush?.(frame.data as Msg)
        }
      } catch {
        // ignore bad frames
      }
    }
    ws.onclose = () => {
      this.onStatus?.(false)
      this.clearTimers()
      if (!this.closedByUser) {
        this.reconnectTimer = window.setTimeout(() => this.open(), 2000)
      }
    }
    ws.onerror = () => {
      ws.close()
    }
  }

  private startHeartbeat(): void {
    this.heartbeatTimer = window.setInterval(() => {
      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
        return
      }
      const frame: Frame = {
        cmd: 'ping',
        req_id: String(this.reqSeq++),
        data: {},
      }
      this.ws.send(JSON.stringify(frame))
    }, 25000)
  }

  private clearTimers(): void {
    if (this.heartbeatTimer != null) {
      window.clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
    if (this.reconnectTimer != null) {
      window.clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
  }
}
