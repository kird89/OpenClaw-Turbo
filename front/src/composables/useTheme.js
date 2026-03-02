import { ref, computed } from 'vue'
import { darkTheme } from 'naive-ui'
import { theme as darkOverrides } from '@/theme/dark'
import { lightTheme as lightOverrides } from '@/theme/light'

const STORAGE_KEY = 'gmssh-theme'

// 读取持久化的主题，默认跟随系统或 dark
function getInitialTheme() {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored === 'light' || stored === 'dark') return stored
    // 如果浏览器偏好亮色，默认亮色
    if (window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches) {
        return 'light'
    }
    return 'dark'
}

const currentTheme = ref(getInitialTheme())

// 初始化 DOM 属性
function applyThemeToDOM(theme) {
    document.documentElement.setAttribute('data-theme', theme)
}

// 初始应用
applyThemeToDOM(currentTheme.value)

export function useTheme() {
    const isDark = computed(() => currentTheme.value === 'dark')

    // Naive UI theme 对象: darkTheme or null (light)
    const naiveTheme = computed(() => isDark.value ? darkTheme : null)

    // Naive UI theme overrides
    const themeOverrides = computed(() => isDark.value ? darkOverrides : lightOverrides)

    function toggleTheme() {
        currentTheme.value = isDark.value ? 'light' : 'dark'
        localStorage.setItem(STORAGE_KEY, currentTheme.value)
        applyThemeToDOM(currentTheme.value)
    }

    function setTheme(theme) {
        currentTheme.value = theme
        localStorage.setItem(STORAGE_KEY, theme)
        applyThemeToDOM(theme)
    }

    return {
        isDark,
        currentTheme,
        naiveTheme,
        themeOverrides,
        toggleTheme,
        setTheme,
    }
}
