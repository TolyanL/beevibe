import '@fontsource-variable/unbounded'
import '@fontsource-variable/manrope'
import '@fontsource/jetbrains-mono/400.css'
import '@fontsource/jetbrains-mono/700.css'
import './assets/styles/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import OpenVue from 'openvue/config'
import ToastService from 'openvue/toastservice'
import ConfirmationService from 'openvue/confirmationservice'

import App from './App.vue'
import router from './router'
import { BeeVibePreset } from './theme/preset'

const app = createApp(App)

app.use(createPinia())
app.use(OpenVue, {
  theme: {
    preset: BeeVibePreset,
    options: {
      darkModeSelector: '.dark',
      cssLayer: {
        name: 'openvue',
        order: 'theme, base, openvue, components, utilities',
      },
    },
  },
  ripple: true,
})
app.use(ToastService)
app.use(ConfirmationService)
app.use(router)

app.mount('#app')
