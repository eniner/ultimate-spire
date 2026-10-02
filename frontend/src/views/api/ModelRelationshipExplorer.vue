<template>
  <content-area style="padding: 0px !important">
    <div class="row">
      <div :class="selectedModel.model_name ? 'col-5' : 'col-12'">
        <eq-window title="API models">
          <div class="ui-toolbar">
            <div class="ui-stat-line mr-auto">
              <b>{{ modelsWithRelations.length }}</b> models with relationships
            </div>
            <input
              class="form-control form-control-sm"
              v-model="search"
              placeholder="Filter models"
            >
          </div>
          <div style="overflow: auto; max-height: 75vh">
            <table class="eq-table">
              <thead>
              <tr>
                <th>Model</th>
                <th>Table</th>
                <th class="text-right" style="width: 90px">Rels</th>
              </tr>
              </thead>
              <tbody>
              <tr
                v-for="model in filteredModels"
                :key="model.model_name"
                class="task-row"
                :class="{ 'is-selected': selected === model.model_name }"
                @click="selected = model.model_name; draw()"
              >
                <td>{{ model.model_name }}</td>
                <td class="text-muted">{{ model.table }}</td>
                <td class="tabular text-right">{{ model.relationships.length }}</td>
              </tr>
              </tbody>
            </table>
          </div>
        </eq-window>
      </div>
      <div class="col-7" v-if="selectedModel.model_name">
        <eq-window :title="selectedModel.model_name">
          <div class="ui-stat-line mb-3">
            Table <b>{{ selectedModel.table }}</b>
            <span class="evo-dot">·</span>
            <b>{{ selectedModel.relationships.length }}</b> relationships
          </div>
          <table class="eq-table" v-if="selectedModel.relationships.length">
            <thead>
            <tr>
              <th>Relationship</th>
            </tr>
            </thead>
            <tbody>
            <tr v-for="(rel, index) in selectedModel.relationships" :key="index">
              <td class="tabular">{{ rel }}</td>
            </tr>
            </tbody>
          </table>
        </eq-window>
      </div>
    </div>
  </content-area>
</template>

<script>
import EqWindow     from "../../components/eq-ui/EQWindow";
import ContentArea  from "../../components/layout/ContentArea";
import {SpireApi}   from "../../app/api/spire-api";
import {HttpStatus} from "../../app/api/http-status";

export default {
  name: "ModelRelationshipExplorer",
  components: { ContentArea, EqWindow },
  data() {
    return {
      selected: "",
      options: [],
      search: "",
      models: [],
      selectedModel: {}
    }
  },
  computed: {
    modelsWithRelations() {
      return this.models.filter((m) => m.relationships && m.relationships.length > 0)
    },
    filteredModels() {
      const q = (this.search || "").toLowerCase()
      return this.modelsWithRelations.filter((m) => {
        if (!q) {
          return true
        }
        return String(m.model_name).toLowerCase().includes(q) || String(m.table).toLowerCase().includes(q)
      })
    },
  },
  methods: {
    draw() {
      for (let model of this.models) {
        if (this.selected === model.model_name) {
          this.selectedModel = model
        }
      }
    }
  },
  async mounted() {
    const r = await SpireApi.v1().get('/models')
    if (r.status === HttpStatus.OK) {
      this.models = r.data.sort((a, b) => a.model_name.localeCompare(b.model_name));

      let options = []
      for (let model of this.models) {
        if (model.relationships.length > 0) {
          options.push({
            value: model.model_name,
            text: `${model.model_name}` + (model.relationships.length > 0 ? ` (${model.relationships.length})` : '')
          })
        }
      }

      this.options = options

    }
  }
}
</script>

<style scoped>

</style>
