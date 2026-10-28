import '@fontsource/nunito-sans/latin-400.css'
import '@fontsource/nunito-sans/latin-600.css'
import '@fontsource/nunito-sans/latin-700.css'
import '@fontsource/nunito-sans/latin-800.css'
import '@fontsource/nunito-sans/latin-900.css'
import '@fontsource/nunito-sans/latin-ext-400.css'
import '@fontsource/nunito-sans/latin-ext-700.css'
import { createApp } from 'vue'
import './assets/bundle.css'
import './theme.css'
import App from './App.vue'
import { applyLocale } from './i18n'

applyLocale()
createApp(App).mount('#app')
