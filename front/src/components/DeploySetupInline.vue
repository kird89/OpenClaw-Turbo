<template>
  <DeploySetup :deploy-mode="deployMode" @back="$emit('back')" />
</template>

<script setup>
import DeploySetup from '@/views/DeploySetup.vue'

defineProps({
  deployMode: { type: String, default: 'docker' }
})
defineEmits(['back'])
</script>
