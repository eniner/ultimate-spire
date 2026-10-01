import {EventBus} from "./event-bus/event-bus"

const STORAGE_KEY = "spire-privacy-mode"

export class Privacy {
  static isOn(): boolean {
    return localStorage.getItem(STORAGE_KEY) === "1"
  }

  static apply() {
    document.documentElement.classList.toggle("spire-privacy", this.isOn())
    EventBus.$emit("SPIRE_PRIVACY", this.isOn())
  }

  static setOn(on: boolean) {
    localStorage.setItem(STORAGE_KEY, on ? "1" : "0")
    this.apply()
  }

  static toggle(): boolean {
    this.setOn(!this.isOn())
    return this.isOn()
  }

  static init() {
    this.apply()
  }
}
