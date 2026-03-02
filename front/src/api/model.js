import gm from '@/utils/gmssh'

export function getModelsConfig() {
    return gm.request('getModelsConfig')
}

export function saveModelsConfig(params) {
    return gm.request('saveModelsConfig', params)
}
