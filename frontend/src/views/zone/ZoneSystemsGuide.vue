<template>
  <content-area>
    <eq-window title="Zone systems guide">
      <demo-zc-pack-notice />

      <div class="ui-toolbar">
        <router-link class="btn btn-sm btn-dark" :to="ROUTE.ZONE_CONTROLLER">Zone Controller</router-link>
        <router-link class="btn btn-sm btn-dark" :to="ROUTE.ZONE_CONTROLLER_BUILDER">Builder</router-link>
        <router-link class="btn btn-sm btn-dark" :to="ROUTE.ZONE_CONTROLLER_FACTORY">Tier Factory</router-link>
        <router-link class="btn btn-sm btn-dark" :to="ROUTE.CONTENT_FACTORY">Content Factory</router-link>
        <button
          v-for="s in sections"
          :key="s.id"
          type="button"
          class="btn btn-sm"
          :class="section === s.id ? 'btn-primary' : 'btn-dark'"
          @click="go(s.id)"
        >
          {{ s.label }}
        </button>
      </div>

      <p class="zc-copy">
        These tools split work: Content Factory mints new PEQ rows. Zone Controller JSON
        scales and rolls loot on live spawns. Builder creates the JSON folders.
        Tier Factory stamps one recipe onto many zones. Apply reloads the Perl controller.
        Always Preview, then Write. Writing files does not change a popped zone until Apply.
      </p>

      <eq-window v-if="section === 'map'" title="Which tool to use">
        <table class="eq-table" style="width: 100%">
          <thead>
          <tr><th>I want to…</th><th>Use</th><th>Writes</th></tr>
          </thead>
          <tbody>
          <tr>
            <td>Turn a PEQ zone on for the controller (blank kit)</td>
            <td><router-link :to="ROUTE.ZONE_CONTROLLER_BUILDER">Builder → Create / clone</router-link></td>
            <td><code>ultimatedata/&lt;id&gt;/&lt;id&gt;_{mob,loot,item}.json</code></td>
          </tr>
          <tr>
            <td>Change trash / boss / raid HP and hits</td>
            <td><router-link :to="ROUTE.ZONE_CONTROLLER_BUILDER">Builder → Apply tier</router-link> or open the zone in Zone Controller</td>
            <td>mob JSON <code>basedata</code></td>
          </tr>
          <tr>
            <td>Mark a named mob as boss / raid</td>
            <td><router-link :to="ROUTE.ZONE_CONTROLLER_BUILDER">Builder → Add custom mobs</router-link></td>
            <td>mob JSON <code>custom</code></td>
          </tr>
          <tr>
            <td>Make new items / loot / NPC types / spell sets</td>
            <td><router-link :to="ROUTE.CONTENT_FACTORY">Content Factory</router-link></td>
            <td>PEQ tables at 800000+ (NPC-castable spells 50000–65535)</td>
          </tr>
          <tr>
            <td>Stamp the same tier onto many zones</td>
            <td><router-link :to="ROUTE.ZONE_CONTROLLER_FACTORY">Tier Factory</router-link></td>
            <td>JSON + optional cloned items</td>
          </tr>
          <tr>
            <td>See it in-game</td>
            <td><router-link :to="ROUTE.ZONE_CONTROLLER">Zone Controller Apply</router-link> or hail the NPC</td>
            <td><code>_spire_commands/&lt;id&gt;.json</code></td>
          </tr>
          </tbody>
        </table>
        <p class="zc-hint mt-3">
          Do not invent items inside Zone Controller JSON and expect them to exist in PEQ.
          Mint the item first, then point loot tables at that ID.
          <code>-1</code> in basedata means leave the vanilla NPC stat alone.
        </p>
      </eq-window>

      <eq-window v-else-if="section === 'zone'" title="Create a zone kit">
        <p class="zc-copy">
          This does not create a new EverQuest zone. It creates controller JSON for an existing PEQ zone id
          (Blackburrow is 17). The in-game NPC named <code>zone_controller</code> reads that JSON.
        </p>
        <ol class="guide-steps">
          <li>
            Confirm the Perl controller is installed
            (<a :href="packUrl" target="_blank" rel="noopener">starter pack</a> /
            <a :href="filesUrl" target="_blank" rel="noopener">where files go</a>).
            NPC <code>2000986</code> must be named <code>zone_controller</code>.
          </li>
          <li>
            Open
            <router-link :to="ROUTE.ZONE_CONTROLLER_BUILDER">Builder → Create / clone</router-link>.
          </li>
          <li>
            Source = <b>Blank template</b> for a clean zone, or <b>Clone a finished zone kit</b> to copy
            basedata / loot / custom from a source zone.
          </li>
          <li>
            Enter the PEQ zone id(s), for example <code>17</code>. Missing-zone pills add ids for you.
          </li>
          <li>
            Leave <b>basedata</b> and <b>ignore</b> checked. Uncheck loot / items / custom if you want empty lists.
            Preview, then <b>Write JSON</b>.
          </li>
          <li>
            Or skip Builder: pop the zone, hail the controller (or <code>!zc refreshzonedata</code>).
            Missing files are copied from <code>ultimatedata/templates</code>.
          </li>
          <li>
            Open
            <router-link :to="ROUTE.ZONE_CONTROLLER">Zone Controller</router-link>
            and click the new zone. Set trash / boss / raid numbers, or use
            <router-link :to="ROUTE.ZONE_CONTROLLER_BUILDER">Builder → Apply tier</router-link>.
          </li>
          <li>
            Name your bosses:
            <router-link :to="ROUTE.ZONE_CONTROLLER_BUILDER">Builder → Add custom mobs</router-link>
            with the exact in-game clean name and type <code>boss</code> or <code>raid</code>.
            Everything else stays trash.
          </li>
          <li>
            Apply live (see the Apply tab). Until then the zone still uses vanilla stats.
          </li>
        </ol>
        <p class="zc-hint">
          Ignore list should include <code>zone controller</code> so the script does not buff itself.
        </p>
      </eq-window>

      <eq-window v-else-if="section === 'items'" title="Create items and loot">
        <p class="zc-copy">
          New gear is cloned from existing PEQ items into the reserved 800000+ range.
          Zone Controller loot JSON only stores table names and chances — the item row must exist first.
        </p>
        <ol class="guide-steps">
          <li>
            Open <router-link :to="ROUTE.CONTENT_FACTORY">Content Factory → IDs</router-link>
            and Refresh. Next item id is <code>max(800000, max+1)</code>.
          </li>
          <li>
            Optional: <b>Census</b> a source zone, uncheck anything you do not want, then
            <b>Fill kit from this zone</b>. Default target is trash; merchants are skipped.
          </li>
          <li>
            Open <b>Item kit</b>. Each cell is a <em>source</em> item id for that class and slot.
            Set a prefix (<code>T1</code>), leave start id 0 unless you need a fixed range.
            Search items, click a result to set Fill value, then fill a class or slot.
          </li>
          <li>
            Give the kit a recipe id (<code>classic-t1</code>) if you will stamp it later.
            <b>Preview</b>, then <b>Write items</b>. That mints new <code>items</code> rows.
          </li>
          <li>
            Open <b>Loot</b>. <b>Use last kit items</b> for the trash table.
            Add a named table only if bosses should drop different gear.
            Attach with zone id (uses the target filter) or explicit NPC ids.
          </li>
          <li>
            <b>Attach</b> can also stock a merchant with the last kit, or set
            <code>loottable_id</code> on the current target NPCs.
          </li>
          <li>
            <b>Test → Give</b> puts minted items on a character so you can inspect them in-game.
          </li>
          <li>
            If you cloned click/proc spells, tick <b>Export spells_us.txt</b> (Spells tab or Pipeline)
            so the client sees the new ids.
          </li>
          <li>
            Open <router-link :to="ROUTE.ZONE_CONTROLLER_FACTORY">Tier Factory</router-link>
            if a recipe was saved. Stamp or ladder that kit onto more zones.
            Ladder mints a new id range so T2 is not wearing T1 loot.
          </li>
        </ol>
        <p class="zc-hint">
          One-shot alternative: Census → fill kit → <b>Pipeline</b> with Use kit + Build trash loot.
          Still Preview first.
        </p>
      </eq-window>

      <eq-window v-else-if="section === 'npcs'" title="Create or retarget NPCs">
        <p class="zc-copy">
          Two different jobs. Scaling a gnoll’s HP is Zone Controller JSON.
          Making a new <code>npc_types</code> row is Content Factory clones at 800000+.
          Do not bulk-edit vanilla PEQ NPC ids.
        </p>
        <h4>A. Scale the NPCs that already spawn</h4>
        <ol class="guide-steps">
          <li>
            Create the zone kit first (New zone tab) so <code>basedata</code> exists.
          </li>
          <li>
            Trash uses <code>basedata.trash</code> automatically. You do not list every gnoll.
          </li>
          <li>
            Named mobs: <router-link :to="ROUTE.ZONE_CONTROLLER_BUILDER">Builder → Add custom mobs</router-link>
            (or the zone’s Custom list). Type the clean name exactly, type <code>boss</code> or <code>raid</code>.
            Mods of <code>-1</code> mean “use that type’s basedata only.”
          </li>
          <li>
            Put vendors, forges, and the controller on the <b>ignore</b> list.
          </li>
          <li>
            Apply <code>refreshzonedata</code> then <code>rebuffzone</code> so live HP updates.
          </li>
        </ol>
        <h4>B. Mint new NPC types (clones)</h4>
        <ol class="guide-steps">
          <li>
            <router-link :to="ROUTE.CONTENT_FACTORY">Content Factory → Census</router-link>
            the zone. Uncheck merchants and any NPC you must not clone.
          </li>
          <li>
            Keep Target filter on <b>trash</b> unless you mean to clone named too.
          </li>
          <li>
            Open <b>Pipeline</b>. Tick <b>Clone NPCs (800000+)</b> and
            <b>Retarget spawnentry to clones</b>. Vanilla rows stay; spawn2 points at the clones.
          </li>
          <li>
            <b>NPC spells</b>: clone a list (Default Magician, etc.). Tick NPC-castable so
            spell ids stay ≤ 65535. Attach the new set to the same target.
          </li>
          <li>
            Preview the pipeline, then Write. Catalog shows the new ids and what uses them.
          </li>
          <li>
            <b>Test → Pawn</b> spawns a <code>SPIRE_TEST_</code> NPC at the zone safe point
            so you can hit it without walking the zone.
          </li>
          <li>
            Zone Controller basedata still applies to clones by name / type after Apply.
          </li>
        </ol>
      </eq-window>

      <eq-window v-else-if="section === 'pipeline'" title="One pipeline (whole tier)">
        <p class="zc-copy">
          Use this when you already know the zone, the kit, and the spell sources.
          Order is fixed: spells → kit → NPC clones → trash/named loot → dual spell sets →
          merchant → client export → reload → probe.
        </p>
        <ol class="guide-steps">
          <li>
            Census the zone. Uncheck named / merchants you do not want mutated.
          </li>
          <li>
            Fill the item kit (or import from census).
          </li>
          <li>
            Optionally pick trash and named spell-set source ids.
          </li>
          <li>
            Open <router-link :to="ROUTE.CONTENT_FACTORY">Content Factory → Pipeline</router-link>.
            Set prefix, step, recipe id, and the zone id.
          </li>
          <li>
            Tick only the stages you want. First-time tier: kit + loot + clone NPCs + retarget + probe.
          </li>
          <li>
            Preview. Read lint. Fix empty kit cells or missing spell ids before Write.
          </li>
          <li>
            Write. If a recipe id was set, open it in
            <router-link :to="ROUTE.ZONE_CONTROLLER_FACTORY">Tier Factory</router-link>
            to stamp more zones.
          </li>
          <li>
            Undo tab can roll back a journaled run if the probe fails.
          </li>
        </ol>
      </eq-window>

      <eq-window v-else-if="section === 'apply'" title="Apply so the zone actually changes">
        <p class="zc-copy">
          JSON on disk is not live memory. The popped <code>zone_controller</code> NPC must reload.
        </p>
        <ol class="guide-steps">
          <li>
            Pop the zone (someone in it, or boot it from Apply).
          </li>
          <li>
            Confirm the invisible NPC named <code>zone_controller</code> exists
            (Tier Factory → Validate, or <code>!togglevis</code> in-game).
          </li>
          <li>
            In <router-link :to="ROUTE.ZONE_CONTROLLER">Zone Controller</router-link> or
            <router-link :to="ROUTE.ZONE_CONTROLLER_FACTORY">Tier Factory → Apply</router-link>
            queue <code>refreshzonedata</code> then <code>rebuffzone</code> (or <code>applyallchanges</code>).
          </li>
          <li>
            A popped controller reads <code>ultimatedata/_spire_commands/&lt;zoneid&gt;.json</code>
            within about 2 seconds and writes <code>&lt;zoneid&gt;.result.json</code>.
          </li>
          <li>
            Empty popped zones can be rebooted so they pick up quests. Occupied zones stay up;
            use the command queue instead of a blind reboot.
          </li>
          <li>
            Fallback from any zone if <code>global_player.pl</code> has the snippet:
            <code>!initdata 17</code>. Or hail the controller and use <code>!zc refreshzonedata</code>.
          </li>
          <li>
            New item / spell / npc_types rows also need a world item/spell reload or zone repop
            so the client and zone process see the PEQ ids. Pipeline can tick that for you.
          </li>
        </ol>
      </eq-window>

      <eq-window v-else title="IDs, files, and rules">
        <table class="eq-table" style="width: 100%">
          <thead>
          <tr><th>Thing</th><th>Where it lives</th></tr>
          </thead>
          <tbody>
          <tr>
            <td>Controller script</td>
            <td><code>quests/global/zone_controller.pl</code></td>
          </tr>
          <tr>
            <td>Zone kit</td>
            <td><code>quests/global/ultimatedata/&lt;id&gt;/&lt;id&gt;_{mob,loot,item}.json</code></td>
          </tr>
          <tr>
            <td>Blank defaults</td>
            <td><code>quests/global/ultimatedata/templates/_*.json</code></td>
          </tr>
          <tr>
            <td>Apply queue</td>
            <td><code>quests/global/ultimatedata/_spire_commands/&lt;id&gt;.json</code></td>
          </tr>
          <tr>
            <td>Recipes / drafts / undo</td>
            <td><code>_spire_recipes</code>, <code>_spire_drafts</code>, <code>_spire_runs</code></td>
          </tr>
          <tr>
            <td>Minted items / NPC clones / loot tables</td>
            <td>PEQ <code>items</code>, <code>npc_types</code>, <code>loottable</code> — ids ≥ 800000</td>
          </tr>
          <tr>
            <td>NPC-castable spells</td>
            <td>PEQ <code>spells_new</code> — 50000–65535 only</td>
          </tr>
          </tbody>
        </table>
        <ul class="guide-rules">
          <li>Preview before every Write.</li>
          <li>Target filter defaults to trash and skips merchants so named and vendors are not bulk-mutated.</li>
          <li>Clone instead of editing vanilla PEQ rows.</li>
          <li>Custom mob names must match the in-game clean name.</li>
          <li>Sage / Lantern / 2D maps are not this pack — they use the EQ client and first-launch assets.</li>
        </ul>
      </eq-window>
    </eq-window>
  </content-area>
</template>

<script>
import EqWindow from "../../components/eq-ui/EQWindow"
import ContentArea from "../../components/layout/ContentArea"
import DemoZcPackNotice from "../../components/DemoZcPackNotice"
import {ROUTE} from "@/routes"
import {demoSiteUrl, isDemoMode} from "@/app/demo-mode"

const SECTIONS = [
  {id: "map", label: "Map"},
  {id: "zone", label: "New zone"},
  {id: "items", label: "Items"},
  {id: "npcs", label: "NPCs"},
  {id: "pipeline", label: "Pipeline"},
  {id: "apply", label: "Apply live"},
  {id: "files", label: "Files / IDs"},
]

export default {
  name: "ZoneSystemsGuide",
  components: {ContentArea, EqWindow, DemoZcPackNotice},
  data() {
    return {
      ROUTE,
      sections: SECTIONS,
    }
  },
  computed: {
    section() {
      const raw = String((this.$route.query && this.$route.query.section) || "map")
      return SECTIONS.some((s) => s.id === raw) ? raw : "map"
    },
    packUrl() {
      return isDemoMode() ? demoSiteUrl("zone-controller-starter.zip") : "https://eniner.github.io/ultimate-spire-demo/zone-controller-starter.zip"
    },
    filesUrl() {
      return isDemoMode() ? demoSiteUrl("#zone-controller") : "https://eniner.github.io/ultimate-spire-demo/#zone-controller"
    },
  },
  methods: {
    go(id) {
      if (id === this.section) {
        return
      }
      this.$router.replace({path: ROUTE.ZONE_CONTROLLER_GUIDE, query: {section: id}})
    },
  },
}
</script>

<style scoped>
.zc-copy, .zc-hint {
  color: var(--text-muted, #b7c0cc);
  margin-bottom: 12px;
}

.guide-steps {
  color: var(--text-muted, #b7c0cc);
  line-height: 1.65;
  padding-left: 22px;
  margin: 0 0 16px;
}

.guide-steps li {
  margin-bottom: 10px;
}

.guide-steps a,
.eq-table a {
  color: var(--text, #e7eaef);
  text-decoration: underline;
}

.guide-rules {
  color: var(--text-muted, #b7c0cc);
  line-height: 1.6;
  margin: 16px 0 0;
  padding-left: 20px;
}

h4 {
  margin: 18px 0 8px;
  font-size: 15px;
}
</style>
