<template>
  <div class="peq-hub">
    <eq-window title="PEQ Editors" class="mb-3 p-3">
      <p class="mb-2 peq-hub-lead">
        Every PHP editor tab from peqphpeditor, mapped onto local Ultimate Spire.
        Dedicated pages stay where they are. Missing ones open a table editor against the same peq tables.
      </p>
      <div class="peq-hub-legend">
        <span class="peq-pill peq-pill-dedicated">Full page</span>
        <span class="peq-pill peq-pill-table">Table editor</span>
      </div>
    </eq-window>

    <div v-for="group in groups" :key="group.id" class="mb-3">
      <eq-window :title="group.title" class="p-3">
        <div class="peq-hub-grid">
          <router-link
            v-for="editor in editorsIn(group.id)"
            :key="editor.id"
            :to="editor.to"
            :class="'peq-hub-card peq-hub-card-' + editor.status"
          >
            <div class="peq-hub-card-top">
              <strong>{{ editor.title }}</strong>
              <span :class="'peq-pill peq-pill-' + editor.status">{{ statusLabel(editor.status) }}</span>
            </div>
            <div class="peq-hub-card-php">php: {{ editor.php }}</div>
            <p>{{ editor.blurb }}</p>
          </router-link>
        </div>
      </eq-window>
    </div>
  </div>
</template>

<script>
import EqWindow from "../../components/eq-ui/EQWindow"
import { PEQ_EDITORS, PEQ_EDITOR_GROUPS } from "../../app/peq-editors"

export default {
  name: "EditorsHub",
  components: { EqWindow },
  data() {
    return {
      editors: PEQ_EDITORS,
      groups: PEQ_EDITOR_GROUPS,
    }
  },
  methods: {
    editorsIn(group) {
      return this.editors.filter((e) => e.group === group)
    },
    statusLabel(status) {
      if (status === "dedicated") {
        return "Full page"
      }
      if (status === "table") {
        return "Table editor"
      }
      return "Not built yet"
    },
  },
}
</script>

<style scoped>
.peq-hub-lead {
  color: inherit;
  opacity: 0.85;
}
.peq-hub-legend {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.peq-hub-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 10px;
}
.peq-hub-card {
  display: block;
  padding: 12px;
  border-radius: 8px;
  border: 1px solid var(--border, rgba(255, 255, 255, 0.1));
  background: var(--surface-2, rgba(18, 22, 27, 0.45));
  color: inherit;
  text-decoration: none;
}
.peq-hub-card:hover {
  border-color: rgba(201, 162, 74, 0.55);
  text-decoration: none;
  color: inherit;
}
.peq-hub-card-soon {
  opacity: 0.65;
}
.peq-hub-card-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}
.peq-hub-card-php {
  font-size: 11px;
  opacity: 0.6;
  margin-bottom: 6px;
}
.peq-hub-card p {
  margin: 0;
  font-size: 13px;
  opacity: 0.85;
}
.peq-pill {
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  padding: 2px 6px;
  border-radius: 999px;
  white-space: nowrap;
}
.peq-pill-dedicated {
  background: #2d5a3d;
}
.peq-pill-table {
  background: #3d4d72;
}
.peq-pill-soon {
  background: #4a4033;
}
</style>
