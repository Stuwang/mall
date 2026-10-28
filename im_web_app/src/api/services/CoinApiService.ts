import type { HttpClient } from '@/api/http/HttpClient'
import type {
  CoinBalanceReply,
  CreateRechargeRequest,
  GiftListReply,
  LedgerReply,
  LedgerRequest,
  PayRechargeRequest,
  RechargeOrder,
  RechargeOrderRequest,
  SendGiftReply,
  SendGiftRequest,
} from '@/api/types'

export class CoinApiService {
  private readonly http: HttpClient

  constructor(http: HttpClient) {
    this.http = http
  }

  async balance(): Promise<CoinBalanceReply> {
    const reply = await this.http.post<CoinBalanceReply>('/im/api/v1/coin/balance', {})
    // encoding/json 会省略 0，没有账户时余额字段不出现。
    return {
      balance: reply.balance ?? 0,
      coins_per_yuan: reply.coins_per_yuan ?? 0,
    }
  }

  ledger(request: LedgerRequest = {}): Promise<LedgerReply> {
    return this.http.post<LedgerReply>('/im/api/v1/coin/ledger', request)
  }

  listGifts(): Promise<GiftListReply> {
    return this.http.post<GiftListReply>('/im/api/v1/gift/list', {})
  }

  async sendGift(request: SendGiftRequest): Promise<SendGiftReply> {
    const reply = await this.http.post<SendGiftReply>('/im/api/v1/gift/send', request)
    return {
      ...reply,
      balance: reply.balance ?? 0,
    }
  }

  createRecharge(request: CreateRechargeRequest): Promise<RechargeOrder> {
    return this.http.post<RechargeOrder>('/im/api/v1/coin/recharge/create', request)
  }

  payRecharge(request: PayRechargeRequest): Promise<RechargeOrder> {
    return this.http.post<RechargeOrder>('/im/api/v1/coin/recharge/pay', request)
  }

  mockPaySuccess(request: RechargeOrderRequest): Promise<RechargeOrder> {
    return this.http.post<RechargeOrder>('/im/api/v1/coin/recharge/mock_pay_success', request)
  }

  getRecharge(request: RechargeOrderRequest): Promise<RechargeOrder> {
    return this.http.post<RechargeOrder>('/im/api/v1/coin/recharge/get', request)
  }
}
