<template>
  <content-area>
    <eq-window title="Ultimate Spire changelog">
      <p class="zc-copy">
        Notes for this build. Generate copies markdown you can paste into
        <code>CHANGELOG.md</code> or a GitHub release.
      </p>
      <div class="ui-toolbar">
        <button
          type="button"
          class="btn btn-sm"
          :class="!focus ? 'btn-primary' : 'btn-dark'"
          @click="focus = ''"
        >
          All
        </button>
        <button
          v-for="release in releases"
          :key="release.version"
          type="button"
          class="btn btn-sm"
          :class="focus === release.version ? 'btn-primary' : 'btn-dark'"
          @click="focus = release.version"
        >
          {{ release.version }}
        </button>
        <button type="button" class="btn btn-sm btn-dark" @click="generate('notes')">
          Generate notes
        </button>
        <button type="button" class="btn btn-sm btn-dark" @click="generate('release')">
          Generate release body
        </button>
      </div>
      <b-alert v-if="copied" show variant="success">Copied to clipboard.</b-alert>
    </eq-window>

    <eq-window
      v-for="release in visible"
      :key="release.version"
      class="mt-3"
      :title="releaseTitle(release)"
    >
      <ul class="guide-steps">
        <li v-for="(item, index) in release.items" :key="release.version + '-' + index">
          <b>{{ item.area }}</b>
          {{ item.text }}
        </li>
      </ul>
    </eq-window>
  </content-area>
</template>

<script>
import EqWindow from "@/components/eq-ui/EQWindow"
import ContentArea from "@/components/layout/ContentArea"
import ClipBoard from "@/app/clipboard/clipboard"
import {Notify} from "@/app/Notify"
import {ULTIMATE_CHANGELOG, formatUltimateChangelog, formatUltimateReleaseNotes} from "@/app/ultimate-changelog"

export default {
  name: "Home",
  components: {EqWindow, ContentArea},
  data() {
    return {
      releases: ULTIMATE_CHANGELOG.filter((release) => release.items.length),
      focus: "",
      copied: false,
    }
  },
  computed: {
    visible() {
      if (!this.focus) {
        return this.releases
      }
      return this.releases.filter((release) => release.version === this.focus)
    },
  },
  methods: {
    releaseTitle(release) {
      return release.date ? release.version + " — " + release.date : release.version
    },
    generate(kind) {
      const text = kind === "release" ? formatUltimateReleaseNotes() : formatUltimateChangelog()
      ClipBoard.copyFromText(text)
      this.copied = true
      Notify.toast("Changelog copied to clipboard")
    },
  },
}
</script>

<style scoped>
.zc-copy {
  color: var(--text-muted, #b7c0cc);
  margin-bottom: 12px;
}

.guide-steps {
  color: var(--text-muted, #b7c0cc);
  line-height: 1.65;
  padding-left: 22px;
  margin: 0;
}

.guide-steps li {
  margin-bottom: 10px;
}

.ui-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
</style>
