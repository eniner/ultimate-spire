import {EventBus} from "@/app/event-bus/event-bus"

export type ThemeMode = "classic" | "light" | "dark"

const STORAGE_KEY = "spire-theme-mode"
const media       = window.matchMedia("(prefers-color-scheme: dark)")

export class Theme {
  static getMode(): ThemeMode {
    const m = localStorage.getItem(STORAGE_KEY)
    if (m === "classic" || m === "light" || m === "dark") {
      return m
    }
    if (m === "system") {
      return media.matches ? "dark" : "classic"
    }
    return "dark"
  }

  static isClassic(): boolean {
    return this.getMode() === "classic"
  }

  static resolved(): "light" | "dark" {
    const m = this.getMode()
    if (m === "classic" || m === "light") {
      return "light"
    }
    return "dark"
  }

  static apply() {
    const root = document.documentElement
    const mode = this.getMode()
    if (mode === "classic") {
      root.classList.remove("spire-modern")
      root.removeAttribute("data-theme")
    } else {
      root.classList.add("spire-modern")
      root.setAttribute("data-theme", mode)
    }
    EventBus.$emit("SPIRE_THEME", mode)
  }

  static setMode(mode: ThemeMode) {
    localStorage.setItem(STORAGE_KEY, mode)
    this.apply()
  }

  static cycle(): ThemeMode {
    const order: ThemeMode[] = ["classic", "light", "dark"]
    const next               = order[(order.indexOf(this.getMode()) + 1) % order.length]
    this.setMode(next)
    return next
  }

  static init() {
    this.apply()
    media.addEventListener("change", () => {
      if (localStorage.getItem(STORAGE_KEY) === "system") {
        this.apply()
      }
    })
  }
}
