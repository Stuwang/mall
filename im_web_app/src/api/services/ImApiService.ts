import type { HttpClient } from '@/api/http/HttpClient'
import type {
  HistoryReply,
  HistoryRequest,
  ListConvReply,
  ListConvRequest,
  ListOnlineReply,
  SendReply,
  SendRequest,
} from '@/api/types'

export class ImApiService {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  listOnline(): Promise<ListOnlineReply> {
    return this.http.post<ListOnlineReply>('/im/api/v1/user/online', {})
  }

  listConv(request: ListConvRequest = {}): Promise<ListConvReply> {
    return this.http.post<ListConvReply>('/im/api/v1/conv/list', request)
  }

  history(request: HistoryRequest): Promise<HistoryReply> {
    return this.http.post<HistoryReply>('/im/api/v1/msg/history', request)
  }

  send(request: SendRequest): Promise<SendReply> {
    return this.http.post<SendReply>('/im/api/v1/msg/send', request)
  }
}
