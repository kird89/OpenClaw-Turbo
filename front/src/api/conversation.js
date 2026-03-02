import gm from '@/utils/gmssh'

/**
 * 获取 OpenClaw WS 连接信息 (wsUrl + token)
 */
export function getClawWsInfo() {
    return gm.request('getClawWsInfo')
}

export function getWsProxyStatus() {
    return gm.request('getWsProxyStatus')
}

export function toggleWsProxy(params) {
    return gm.request('toggleWsProxy', params)
}
