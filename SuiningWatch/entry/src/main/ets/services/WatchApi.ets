import { http } from '@kit.NetworkKit';
import { WatchConfig, Credentials, DeviceReply, BindRequest, HeartbeatRequest,
  BatchRequest, SOSRequest, LoginRequest, DeviceRequestError, WATCH_APP_VERSION, WATCH_PROTOCOL_VERSION, verifyHandshake } from '../model/Protocol';

// A single request has a bounded lifetime; requests never log the device secret.
export class WatchApi {
  private active: Set<http.HttpRequest> = new Set<http.HttpRequest>();
  private generation: number = 0;
  cancel(): void {
    this.generation++;
    for (const request of this.active) {
      try { request.destroy(); } catch (error) { /* Already closed by the platform. */ }
    }
    this.active.clear();
  }
  private async post(config: WatchConfig, path: string,
    body: Credentials | LoginRequest | BindRequest | HeartbeatRequest | BatchRequest | SOSRequest): Promise<DeviceReply> {
    const request = http.createHttp();
    const generation = this.generation;
    this.active.add(request);
    try {
      let response: http.HttpResponse;
      try { response = await request.request(config.server + '/api/v1/device/' + path, {
        method: http.RequestMethod.POST,
        header: { 'Content-Type': 'application/json' },
        extraData: JSON.stringify(body),
        usingCache: false,
        expectDataType: http.HttpDataType.STRING,
        connectTimeout: 5000, readTimeout: 10000
      }); } catch (error) {
        if (generation !== this.generation) { throw new DeviceRequestError('CANCELLED', '操作已取消'); }
        throw new DeviceRequestError('NETWORK_ERROR', '网络不可用或请求超时，待上传记录已保留');
      }
      if (generation !== this.generation) { throw new DeviceRequestError('CANCELLED', '操作已取消'); }
      let reply: DeviceReply;
      try { reply = JSON.parse(response.result as string) as DeviceReply; }
      catch (error) { throw new DeviceRequestError('PROTOCOL_ERROR', '服务响应格式无效，待上传记录已保留', response.responseCode); }
      if (!reply || typeof reply !== 'object' || Array.isArray(reply)) {
        throw new DeviceRequestError('PROTOCOL_ERROR', '服务响应格式无效，待上传记录已保留', response.responseCode);
      }
      if (response.responseCode < 200 || response.responseCode >= 300 || reply.returnCode !== '1111') {
        const message = reply.message || '服务返回 HTTP ' + response.responseCode;
        throw new DeviceRequestError(reply.code || 'HTTP_ERROR', message + (reply.reference ? '（编号：' + reply.reference + '）' : ''), response.responseCode);
      }
      return reply;
    } finally {
      if (this.active.delete(request)) {
        try { request.destroy(); } catch (error) { /* Cleanup must not replace a response or its original error. */ }
      }
    }
  }
  async login(config: WatchConfig, secret: string): Promise<DeviceReply> {
    const reply = await this.post(config, 'login', { deviceNo: config.deviceNo, secret: secret,
      platform: 'harmonyos', protocolVersion: WATCH_PROTOCOL_VERSION });
    verifyHandshake(reply);
    return reply;
  }
  bind(config: WatchConfig, secret: string, code: string): Promise<DeviceReply> {
    return this.post(config, 'bind', { deviceNo: config.deviceNo, secret: secret, code: code.trim().toUpperCase() });
  }
  heartbeat(config: WatchConfig, secret: string): Promise<DeviceReply> {
    return this.post(config, 'heartbeat', { deviceNo: config.deviceNo, secret: secret, appVersion: WATCH_APP_VERSION });
  }
  upload(config: WatchConfig, body: BatchRequest): Promise<DeviceReply> {
    return this.post(config, 'healthBatch', body);
  }
  sos(config: WatchConfig, body: SOSRequest): Promise<DeviceReply> {
    return this.post(config, 'sos', body);
  }
}
