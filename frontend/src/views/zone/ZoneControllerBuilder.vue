<template>
  <content-area>
    <eq-window title="Zone Controller builder">
      <b-alert show variant="danger" v-if="error">
        <i class="fa fa-warning"></i> {{ error }}
      </b-alert>
      <b-alert show variant="info" v-if="plan && plan.dryRun && plan.ok">
        Preview only. {{ (plan.writes || []).length }} file action(s). Nothing written yet.
      </b-alert>
      <b-alert show variant="success" v-if="plan && !plan.dryRun && plan.ok">
        Wrote {{ writtenCount }} file(s). In-game still needs <code>!initdata zoneid</code> or hail refresh.
      </b-alert>

      <div class="ui-toolbar">
        <router-link class="btn btn-sm btn-dark" :to="ROUTE.ZONE_CONTROLLER">All ZC zones</router-link>
        <router-link class="btn btn-sm btn-dark" :to="ROUTE.ZONE_CONTROLLER_FACTORY">Tier Factory</router-link>
        <router-link class="btn btn-sm btn-dark" :to="ROUTE.CONTENT_FACTORY">Content Factory</router-link>
        <router-link class="btn btn-sm btn-dark" :to="ROUTE.ZONES">Zones</router-link>
        <button type="button" class="btn btn-sm" :class="tab === 'create' ? 'btn-primary' : 'btn-dark'" @click="tab = 'create'">Create / clone</button>
        <button type="button" class="btn btn-sm" :class="tab === 'tier' ? 'btn-primary' : 'btn-dark'" @click="tab = 'tier'">Apply tier</button>
        <button type="button" class="btn btn-sm" :class="tab === 'mobs' ? 'btn-primary' : 'btn-dark'" @click="tab = 'mobs'">Add custom mobs</button>
      </div>

      <p class="zc-copy">
        Runtime still belongs to <code>zone_controller.pl</code>: every spawn is trash unless listed in
        <code>custom</code>, then stats are <code>basedata[type]</code> plus <code>_mod</code> / <code>_override</code>,
        and loot rolls named tables from the loot JSON. This page only writes those files.
      </p>

      <eq-window v-if="tab === 'create'" title="Create zone JSON">
        <div class="ui-field-grid">
          <div class="ui-field">
            <label>Source</label>
            <select v-model="create.source" class="form-control form-control-sm">
              <option value="template">Blank template (same as in-game !initdata)</option>
              <option value="clone">Clone a finished zone kit</option>
            </select>
          </div>
          <div class="ui-field" v-if="create.source === 'clone'">
            <label>Clone from</label>
            <select v-model.number="create.cloneFrom" class="form-control form-control-sm">
              <option :value="0">Select a source zone</option>
              <option v-for="z in configured" :key="z.zoneId" :value="z.zoneId">
                {{ z.zoneId }} — {{ zoneLabel(z.zoneId) }}
              </option>
            </select>
          </div>
          <div class="ui-field">
            <label>Display name (info.name)</label>
            <input v-model="create.name" class="form-control form-control-sm" placeholder="Optional">
          </div>
          <div class="ui-field">
            <label>Respawn note</label>
            <input v-model="create.respawn" class="form-control form-control-sm" placeholder="Optional">
          </div>
        </div>

        <div class="ui-field mt-3">
          <label>Target zone IDs</label>
          <textarea v-model="create.idsText" class="form-control form-control-sm" rows="3" placeholder="17, 31, 39 or one per line"></textarea>
          <div class="zc-hint">{{ parseIds(create.idsText).length }} zone(s)</div>
        </div>

        <div class="zc-missing" v-if="missing.length">
          <div class="zc-hint">PEQ zones with no ultimatedata yet (click to add):</div>
          <button
            v-for="z in missing.slice(0, 40)"
            :key="z.zoneidnumber"
            type="button"
            class="zc-pill ui-plain"
            @click="addCreateId(z.zoneidnumber)"
          >
            {{ z.zoneidnumber }} {{ z.short_name }}
          </button>
        </div>

        <div class="zc-checks">
          <label><input type="checkbox" v-model="create.include.basedata"> basedata (tier stats)</label>
          <label><input type="checkbox" v-model="create.include.loot"> loot tables</label>
          <label><input type="checkbox" v-model="create.include.items"> items</label>
          <label><input type="checkbox" v-model="create.include.custom"> custom mobs</label>
          <label><input type="checkbox" v-model="create.include.ignore"> ignore list</label>
          <label><input type="checkbox" v-model="create.include.depop"> depop list</label>
          <label><input type="checkbox" v-model="create.overwrite"> overwrite existing files</label>
        </div>

        <div class="ui-toolbar">
          <button class="btn btn-sm btn-dark" :disabled="busy" @click="runCreate(true)">Preview</button>
          <button class="btn btn-sm btn-primary" :disabled="busy" @click="runCreate(false)">Write JSON</button>
        </div>
      </eq-window>

      <eq-window v-else-if="tab === 'tier'" title="Stamp trash / boss / raid onto zones">
        <p class="zc-copy">
          Copy <code>basedata</code> from a source zone (Blackburrow 71/500k trash is a common kit) and/or type new
          numbers. Check copy loot kit if those drop table IDs should exist in the target too.
        </p>
        <div class="ui-field-grid">
          <div class="ui-field">
            <label>Copy basedata from</label>
            <select v-model.number="tier.fromZone" class="form-control form-control-sm">
              <option :value="0">None — only typed fields</option>
              <option v-for="z in configured" :key="z.zoneId" :value="z.zoneId">
                {{ z.zoneId }} — {{ zoneLabel(z.zoneId) }} ({{ z.customCount }} custom)
              </option>
            </select>
          </div>
          <div class="ui-field">
            <label>Copy loot + items from that zone</label>
            <select v-model="tier.copyLoot" class="form-control form-control-sm">
              <option :value="false">No</option>
              <option :value="true">Yes</option>
            </select>
          </div>
        </div>
        <table class="eq-table mt-3" style="width: 100%">
          <thead>
          <tr>
            <th></th>
            <th>Level</th>
            <th>Max HP</th>
            <th>Min hit</th>
            <th>Max hit</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="kind in ['trash', 'boss', 'raid']" :key="kind">
            <td><b>{{ kind }}</b></td>
            <td><input v-model="tier[kind].level" class="form-control form-control-sm" placeholder="keep"></td>
            <td><input v-model="tier[kind].max_hp" class="form-control form-control-sm" placeholder="keep"></td>
            <td><input v-model="tier[kind].min_hit" class="form-control form-control-sm" placeholder="keep"></td>
            <td><input v-model="tier[kind].max_hit" class="form-control form-control-sm" placeholder="keep"></td>
          </tr>
          </tbody>
        </table>
        <div class="ui-field mt-3">
          <label>Target zone IDs (must already have JSON)</label>
          <textarea v-model="tier.idsText" class="form-control form-control-sm" rows="3" placeholder="17, 31, 59"></textarea>
        </div>
        <div class="ui-toolbar">
          <button class="btn btn-sm btn-dark" :disabled="busy" @click="runTier(true)">Preview</button>
          <button class="btn btn-sm btn-primary" :disabled="busy" @click="runTier(false)">Write JSON</button>
        </div>
      </eq-window>

      <eq-window v-else title="Batch custom mobs">
        <p class="zc-copy">
          Same as <code>!remoteaddcustomtypeboss</code> / trash / raid: name + type, default mods <code>-1</code>.
          After write, hail the controller or <code>!initdata</code> so live memory reloads.
        </p>
        <div class="ui-field-grid">
          <div class="ui-field">
            <label>Type</label>
            <select v-model="mobs.type" class="form-control form-control-sm">
              <option value="trash">trash</option>
              <option value="boss">boss</option>
              <option value="raid">raid</option>
            </select>
          </div>
          <div class="ui-field">
            <label>Static named (repop/depop bosses)</label>
            <select v-model="mobs.static" class="form-control form-control-sm">
              <option value="-1">No</option>
              <option value="1">Yes</option>
            </select>
          </div>
        </div>
        <div class="ui-field mt-3">
          <label>Mob names</label>
          <textarea v-model="mobs.namesText" class="form-control form-control-sm" rows="5" placeholder="Fippy Darkpaw&#10;Lord Elgnub"></textarea>
        </div>
        <div class="ui-field">
          <label>Target zone IDs</label>
          <textarea v-model="mobs.idsText" class="form-control form-control-sm" rows="2"></textarea>
        </div>
        <div class="ui-toolbar">
          <button class="btn btn-sm btn-dark" :disabled="busy" @click="runMobs(true)">Preview</button>
          <button class="btn btn-sm btn-primary" :disabled="busy" @click="runMobs(false)">Write JSON</button>
        </div>
      </eq-window>

      <eq-window v-if="plan" title="Plan" class="mt-3">
        <div v-if="plan.error" class="text-danger mb-2">{{ plan.error }}</div>
        <div v-if="(plan.warnings || []).length" class="mb-2">
          <div v-for="w in plan.warnings" :key="w" class="text-warning">{{ w }}</div>
        </div>
        <table class="eq-table" style="width: 100%">
          <thead>
          <tr>
            <th>Zone</th>
            <th>File</th>
            <th>Action</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="w in (plan.writes || [])" :key="w.relPath + w.action">
            <td class="tabular">{{ w.zoneId }}</td>
            <td><code>{{ w.relPath }}</code></td>
            <td>{{ w.action }}</td>
          </tr>
          </tbody>
        </table>
        <div v-if="(plan.backups || []).length" class="zc-hint mt-2">
          Backed up to <code>ultimatedata/_spire_backups</code>
        </div>
      </eq-window>
    </eq-window>
  </content-area>
</template>

<script>
import EqWindow from "../../components/eq-ui/EQWindow"
import ContentArea from "../../components/layout/ContentArea"
import {ROUTE} from "@/routes"
import {Zones} from "../../app/zones"
import {ZoneControllerApi} from "../../app/zone-controller"

export default {
  name: "ZoneControllerBuilder",
  components: {ContentArea, EqWindow},
  data() {
    return {
      ROUTE,
      error: "",
      busy: false,
      tab: "create",
      plan: null,
      configured: [],
      peqZones: {},
      peqList: [],
      create: {
        source: "template",
        cloneFrom: 0,
        name: "",
        respawn: "",
        idsText: "",
        overwrite: false,
        include: {basedata: true, loot: true, items: true, custom: false, ignore: true, depop: true},
      },
      tier: {
        fromZone: 0,
        copyLoot: false,
        idsText: "",
        trash: {level: "", max_hp: "", min_hit: "", max_hit: ""},
        boss: {level: "", max_hp: "", min_hit: "", max_hit: ""},
        raid: {level: "", max_hp: "", min_hit: "", max_hit: ""},
      },
      mobs: {
        type: "boss",
        static: "-1",
        namesText: "",
        idsText: "",
      },
    }
  },
  computed: {
    missing() {
      const have = {}
      this.configured.forEach((z) => {
        have[z.zoneId] = true
      })
      return (this.peqList || []).filter((z) => z.version === 0 && z.zoneidnumber && !have[z.zoneidnumber])
        .sort((a, b) => a.zoneidnumber - b.zoneidnumber)
    },
    writtenCount() {
      return (this.plan && this.plan.writes || []).filter((w) => w.action === "create" || w.action === "replace").length
    },
  },
  async mounted() {
    try {
      const zones = await Zones.getZones()
      this.peqList = zones || []
      const byId = {}
      this.peqList.forEach((z) => {
        if (z && z.zoneidnumber && !byId[z.zoneidnumber]) {
          byId[z.zoneidnumber] = z
        }
      })
      this.peqZones = byId
      const index = await ZoneControllerApi.list()
      this.configured = index.zones || []
    } catch (e) {
      this.error = (e.response && e.response.data && e.response.data.error) || String(e)
    }
  },
  methods: {
    parseIds(text) {
      return String(text || "").split(/[,\s]+/).map((s) => parseInt(s, 10)).filter((n) => n > 0)
    },
    parseNames(text) {
      return String(text || "").split(/\r?\n/).map((s) => s.trim()).filter(Boolean)
    },
    zoneLabel(id) {
      const z = this.peqZones[id]
      return z && z.long_name ? z.long_name : ("Zone " + id)
    },
    addCreateId(id) {
      const cur = this.parseIds(this.create.idsText)
      if (cur.indexOf(id) === -1) {
        cur.push(id)
      }
      this.create.idsText = cur.join(", ")
    },
    async runCreate(dryRun) {
      await this.run(() => ZoneControllerApi.create({
        zoneIds: this.parseIds(this.create.idsText),
        source: this.create.source,
        cloneFrom: this.create.cloneFrom,
        include: this.create.include,
        name: this.create.name,
        respawn: this.create.respawn,
        overwrite: this.create.overwrite,
        dryRun,
      }))
    },
    async runTier(dryRun) {
      await this.run(() => ZoneControllerApi.applyTier({
        zoneIds: this.parseIds(this.tier.idsText),
        fromZone: this.tier.fromZone,
        trash: this.tier.trash,
        boss: this.tier.boss,
        raid: this.tier.raid,
        copyLoot: this.tier.copyLoot === true || this.tier.copyLoot === "true",
        dryRun,
      }))
    },
    async runMobs(dryRun) {
      await this.run(() => ZoneControllerApi.addMobs({
        zoneIds: this.parseIds(this.mobs.idsText),
        type: this.mobs.type,
        names: this.parseNames(this.mobs.namesText),
        static: this.mobs.static,
        dryRun,
      }))
    },
    async run(fn) {
      this.busy = true
      this.error = ""
      this.plan = null
      try {
        this.plan = await fn()
        if (this.plan && this.plan.error) {
          this.error = this.plan.error
        } else if (this.plan && !this.plan.dryRun) {
          const index = await ZoneControllerApi.list()
          this.configured = index.zones || []
        }
      } catch (e) {
        const status = e.response && e.response.status
        if (status === 404) {
          this.error = "Builder write routes are not on the running backend yet. Restart Spire with start_spire_dev.bat so create/tier/mobs load."
        } else {
          this.error = (e.response && e.response.data && e.response.data.error) || String(e)
        }
        if (e.response && e.response.data && e.response.data.writes) {
          this.plan = e.response.data
        }
      }
      this.busy = false
    },
  },
}
</script>

<style scoped>
.zc-copy, .zc-hint {
  color: var(--text-muted, #b7c0cc);
  margin-bottom: 12px;
}

.zc-checks {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 18px;
  margin: 12px 0;
}

.zc-checks label {
  margin: 0;
}

.zc-missing {
  margin: 10px 0 14px;
}

.zc-pill {
  display: inline-block;
  margin: 0 4px 4px 0;
  padding: 1px 7px;
  border: 1px solid var(--border, #3a4350);
  border-radius: 999px;
  background: var(--surface-2, #12161d);
  color: inherit;
  font-size: 12px;
}

.mt-3 {
  margin-top: 1rem;
}
</style>
