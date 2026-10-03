<template>
  <content-area>
    <eq-window title="Content Factory">
      <b-alert show variant="danger" v-if="error">
        <i class="fa fa-warning"></i> {{ error }}
      </b-alert>
      <b-alert show variant="info" v-if="plan && plan.dryRun && plan.ok">
        Preview only. {{ (plan.dbWrites || []).length }} database write(s), {{ (plan.impact || []).length }} NPC/merchant change(s). Nothing written yet.
      </b-alert>
      <b-alert show variant="success" v-if="plan && !plan.dryRun && plan.ok">
        Wrote {{ plan.kind }}.
        <span v-if="plan.runId"> Run {{ plan.runId }} can be undone.</span>
        <span v-if="plan.recipeId"> Recipe <b>{{ plan.recipeId }}</b> is ready in Tier Factory.</span>
        <span v-if="plan.export && plan.export.spellsPath"> Exported {{ plan.export.spellsPath }}.</span>
      </b-alert>

      <div class="ui-toolbar">
        <router-link class="btn btn-sm btn-dark" :to="ROUTE.ZONE_CONTROLLER_FACTORY">Tier Factory</router-link>
        <router-link class="btn btn-sm btn-dark" :to="ROUTE.ZONE_CONTROLLER">Zone Controller</router-link>
        <router-link class="btn btn-sm btn-dark" :to="ROUTE.ZONE_CONTROLLER_GUIDE">Guide</router-link>
        <button type="button" class="btn btn-sm" :class="tab === 'ids' ? 'btn-primary' : 'btn-dark'" @click="tab = 'ids'">IDs</button>
        <button type="button" class="btn btn-sm" :class="tab === 'census' ? 'btn-primary' : 'btn-dark'" @click="tab = 'census'">Census</button>
        <button type="button" class="btn btn-sm" :class="tab === 'catalog' ? 'btn-primary' : 'btn-dark'" @click="tab = 'catalog'; loadCatalog()">Catalog</button>
        <button type="button" class="btn btn-sm" :class="tab === 'spells' ? 'btn-primary' : 'btn-dark'" @click="tab = 'spells'">Spells</button>
        <button type="button" class="btn btn-sm" :class="tab === 'npc' ? 'btn-primary' : 'btn-dark'" @click="tab = 'npc'">NPC spells</button>
        <button type="button" class="btn btn-sm" :class="tab === 'kit' ? 'btn-primary' : 'btn-dark'" @click="tab = 'kit'">Item kit</button>
        <button type="button" class="btn btn-sm" :class="tab === 'loot' ? 'btn-primary' : 'btn-dark'" @click="tab = 'loot'">Loot</button>
        <button type="button" class="btn btn-sm" :class="tab === 'attach' ? 'btn-primary' : 'btn-dark'" @click="tab = 'attach'">Attach</button>
        <button type="button" class="btn btn-sm" :class="tab === 'pipeline' ? 'btn-primary' : 'btn-dark'" @click="tab = 'pipeline'">Pipeline</button>
        <button type="button" class="btn btn-sm" :class="tab === 'editor' ? 'btn-primary' : 'btn-dark'" @click="tab = 'editor'">Spell set</button>
        <button type="button" class="btn btn-sm" :class="tab === 'test' ? 'btn-primary' : 'btn-dark'" @click="tab = 'test'">Test</button>
        <button type="button" class="btn btn-sm" :class="tab === 'undo' ? 'btn-primary' : 'btn-dark'" @click="tab = 'undo'; loadRuns()">Undo</button>
      </div>

      <p class="zc-copy">
        Target trash by default. Clone NPCs at 800000+ instead of editing vanilla rows.
        Lint and a post-write probe run before anything is treated as done.
      </p>

      <eq-window title="Target filter">
        <p class="zc-copy">Used by census cherry-pick, loot, attach, NPC clones, spell sets, and pipeline. Default is trash, merchants skipped.</p>
        <div class="zc-checks">
          <label v-for="role in roleChoices" :key="role">
            <input type="checkbox" :checked="hasRole(role)" @change="toggleRole(role)"> {{ role }}
          </label>
          <label><input type="checkbox" v-model="target.skipMerchants"> Skip merchants</label>
          <label><input type="checkbox" v-model="target.includeRing"> Include zone ring</label>
        </div>
        <div class="ui-field-grid">
          <div class="ui-field"><label>Name contains</label><input v-model="target.nameContains" class="form-control form-control-sm" placeholder="gnoll"></div>
          <div class="ui-field"><label>Exclude NPC IDs</label><input v-model="target.excludeNpcIds" class="form-control form-control-sm" placeholder="5036, 5040"></div>
          <div class="ui-field"><label>Only NPC IDs</label><input v-model="target.onlyNpcIds" class="form-control form-control-sm" placeholder="leave empty for all matching"></div>
        </div>
        <div class="zc-hint">Draft {{ draftNote }}</div>
      </eq-window>

      <eq-window v-if="tab === 'ids'" title="Reserved ID board">
        <p class="zc-copy">
          Next free ID is <code>max(floor, max+1)</code>. Item clones stay at 800000+.
          NPC-castable spells use the 50000–65535 band.
        </p>
        <div class="ui-toolbar">
          <button type="button" class="btn btn-sm btn-dark" :disabled="busy" @click="loadIds">Refresh</button>
        </div>
        <table class="eq-table" style="width: 100%">
          <thead>
          <tr>
            <th>Table</th>
            <th>Max</th>
            <th>Floor</th>
            <th>Next</th>
            <th>Used ≥ floor</th>
            <th>Note</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="row in (board.rows || [])" :key="row.key">
            <td><b>{{ row.table }}</b>.{{ row.column }}</td>
            <td class="tabular">{{ row.maxId }}</td>
            <td class="tabular">{{ row.floor }}</td>
            <td class="tabular">{{ row.nextId }}</td>
            <td class="tabular">{{ row.reservedUsed }}</td>
            <td>{{ row.note }}</td>
          </tr>
          </tbody>
        </table>
      </eq-window>

      <eq-window v-else-if="tab === 'census'" title="Zone census">
        <p class="zc-copy">PEQ spawn data: roles, faction, neighbors, ground/forage, loot click/proc, and merchants. Uncheck a row to exclude it from writes.</p>
        <div class="ui-field-grid">
          <div class="ui-field"><label>Zone ID or short name</label><input v-model="zoneQuery" class="form-control form-control-sm" placeholder="17 or blackburrow" @keyup.enter="loadCensus"></div>
        </div>
        <div class="ui-toolbar">
          <button type="button" class="btn btn-sm btn-primary" :disabled="busy" @click="loadCensus">Load census</button>
          <button type="button" class="btn btn-sm btn-dark" :disabled="!census.ok" @click="importCensusKit">Fill kit from this zone</button>
          <button type="button" class="btn btn-sm btn-dark" :disabled="!census.ok" @click="useCensusOnly">Use checked as only-IDs</button>
          <button type="button" class="btn btn-sm btn-dark" :disabled="!census.ok" @click="tab = 'pipeline'">Use in pipeline</button>
        </div>
        <div class="zc-hint" v-if="census.ok">
          {{ census.long || census.short }} · {{ (census.counts && census.counts.npcs) || 0 }} NPCs
          ({{ (census.counts && census.counts.trash) || 0 }} trash /
          {{ (census.counts && census.counts.boss) || 0 }} named /
          {{ (census.counts && census.counts.raid) || 0 }} raid /
          {{ (census.counts && census.counts.ignore) || 0 }} ignore) ·
          {{ (census.counts && census.counts.items) || 0 }} loot items ·
          {{ (census.counts && census.counts.spellSets) || 0 }} spell sets ·
          {{ (census.counts && census.counts.merchants) || 0 }} merchants ·
          {{ census.groundSpawns || 0 }} ground · {{ census.forages || 0 }} forage
        </div>
        <div class="zc-missing" v-if="(census.neighbors || []).length">
          Ring:
          <button v-for="n in census.neighbors" :key="n.id" type="button" class="zc-pill ui-plain" @click="openNeighbor(n)">
            {{ n.short }} {{ n.id }}
          </button>
        </div>
        <table v-if="(census.npcs || []).length" class="eq-table mt-3" style="width: 100%">
          <thead><tr><th></th><th>Role</th><th>NPC</th><th>Lvl</th><th>HP</th><th>Pops</th><th>Faction</th><th>Loot</th><th>Spell set</th><th>Casts</th><th>Merchant</th></tr></thead>
          <tbody>
          <tr v-for="n in census.npcs" :key="n.id" :class="{'row-skip': isExcluded(n.id)}">
            <td><input type="checkbox" :checked="!isExcluded(n.id)" @change="toggleExclude(n.id)"></td>
            <td><span class="role-pill" :class="'role-' + (n.role || 'trash')">{{ n.role || "—" }}</span></td>
            <td><router-link :to="'/npc/' + n.id">{{ n.name }}</router-link> <span class="text-muted tabular">{{ n.id }}</span></td>
            <td class="tabular">{{ n.level }}</td>
            <td class="tabular">{{ n.hp || "—" }}</td>
            <td class="tabular">{{ n.pops || "—" }}</td>
            <td>{{ n.faction || n.factionId || "—" }}</td>
            <td class="tabular">{{ n.loottableId || "—" }}</td>
            <td>{{ n.spellSet || n.npcSpellsId || "—" }}</td>
            <td>{{ spellList(n.spells) }}</td>
            <td class="tabular">{{ n.merchantId || "—" }}</td>
          </tr>
          </tbody>
        </table>
        <table v-if="(census.items || []).length" class="eq-table mt-3" style="width: 100%">
          <thead><tr><th>Item</th><th>Slot</th><th>Click</th><th>Proc</th></tr></thead>
          <tbody>
          <tr v-for="it in census.items" :key="it.id">
            <td><router-link :to="'/item/' + it.id">{{ it.name }}</router-link> <span class="text-muted tabular">{{ it.id }}</span></td>
            <td>{{ it.slot }} {{ it.class }}</td>
            <td><router-link v-if="it.click" :to="'/spell/' + it.click">{{ it.clickName || it.click }}</router-link><span v-else>—</span></td>
            <td><router-link v-if="it.proc" :to="'/spell/' + it.proc">{{ it.procName || it.proc }}</router-link><span v-else>—</span></td>
          </tr>
          </tbody>
        </table>
      </eq-window>

      <eq-window v-else-if="tab === 'catalog'" title="Minted catalog">
        <p class="zc-copy">Rows you already minted at 800000+. Uses shows which NPCs, items, or merchants point at them.</p>
        <div class="ui-toolbar">
          <button type="button" class="btn btn-sm" :class="catalogKind === 'items' ? 'btn-primary' : 'btn-dark'" @click="catalogKind = 'items'; loadCatalog()">Items</button>
          <button type="button" class="btn btn-sm" :class="catalogKind === 'spells' ? 'btn-primary' : 'btn-dark'" @click="catalogKind = 'spells'; loadCatalog()">Spells</button>
          <button type="button" class="btn btn-sm" :class="catalogKind === 'loottable' ? 'btn-primary' : 'btn-dark'" @click="catalogKind = 'loottable'; loadCatalog()">Loot</button>
          <button type="button" class="btn btn-sm" :class="catalogKind === 'npc_spells' ? 'btn-primary' : 'btn-dark'" @click="catalogKind = 'npc_spells'; loadCatalog()">NPC sets</button>
        </div>
        <table class="eq-table" style="width: 100%">
          <thead><tr><th>ID</th><th>Name</th><th>Extra</th><th>Used by</th></tr></thead>
          <tbody>
          <tr v-for="row in (catalog.rows || [])" :key="catalogKind + row.id">
            <td class="tabular">{{ row.id }}</td>
            <td>{{ row.name }}</td>
            <td class="text-muted">{{ row.extra }}</td>
            <td>{{ row.useNote }}</td>
          </tr>
          <tr v-if="!(catalog.rows || []).length">
            <td colspan="4" class="text-muted">No reserved rows in this table yet.</td>
          </tr>
          </tbody>
        </table>
      </eq-window>

      <eq-window v-else-if="tab === 'spells'" title="Spell factory">
        <p class="zc-copy">
          Clone source spells and optionally write the new ID onto item click/proc/worn/focus.
          Tick NPC-castable to mint in the 50000–65535 band.
        </p>
        <div class="ui-field-grid">
          <div class="ui-field"><label>Prefix</label><input v-model="spells.prefix" class="form-control form-control-sm" placeholder="T1"></div>
          <div class="ui-field"><label>First new spell ID (0 = next free)</label><input v-model.number="spells.startId" type="number" class="form-control form-control-sm"></div>
          <div class="ui-field"><label>Copies per source</label><input v-model.number="spells.count" type="number" class="form-control form-control-sm"></div>
          <div class="ui-field"><label>Step multiplier</label><input v-model.number="spells.step" type="number" step="0.05" class="form-control form-control-sm"></div>
        </div>
        <div class="zc-checks">
          <label><input type="checkbox" v-model="spells.npcCastable"> NPC-castable (≤65535)</label>
          <label><input type="checkbox" v-model="spells.exportClient"> Export spells_us.txt after write</label>
          <label><input type="checkbox" v-model="spells.scale.effects"> Scale effect values</label>
          <label><input type="checkbox" v-model="spells.scale.max"> Scale max values</label>
          <label><input type="checkbox" v-model="spells.scale.mana"> Scale mana</label>
          <label><input type="checkbox" v-model="spells.scale.duration"> Scale duration</label>
          <label><input type="checkbox" v-model="spells.scale.recast"> Scale recast</label>
        </div>
        <div class="ui-field-grid">
          <div class="ui-field">
            <label>Search spells</label>
            <input v-model="search.spells" class="form-control form-control-sm" placeholder="Complete Heal or 13" @keyup.enter="doSearch('spells')">
          </div>
          <div class="ui-field">
            <label>&nbsp;</label>
            <button type="button" class="btn btn-sm btn-dark" :disabled="busy" @click="doSearch('spells')">Search</button>
          </div>
        </div>
        <div class="zc-missing" v-if="(results.spells || []).length">
          <button v-for="row in results.spells" :key="row.id" type="button" class="zc-pill ui-plain" @click="addSpell(row)">
            {{ row.id }} {{ row.name }}
          </button>
        </div>
        <table class="eq-table mt-3" style="width: 100%">
          <thead><tr><th>Source ID</th><th>Name</th><th></th></tr></thead>
          <tbody>
          <tr v-for="(src, i) in spells.sources" :key="i">
            <td><input v-model.number="src.id" type="number" class="form-control form-control-sm"></td>
            <td><input v-model="src.name" class="form-control form-control-sm"></td>
            <td><button type="button" class="btn btn-sm btn-outline-danger" @click="spells.sources.splice(i, 1)">Remove</button></td>
          </tr>
          <tr v-if="!spells.sources.length">
            <td colspan="3" class="text-muted">Search and click a spell, or add a row.</td>
          </tr>
          </tbody>
        </table>
        <div class="ui-toolbar">
          <button type="button" class="btn btn-sm btn-dark" @click="spells.sources.push({id: 0, name: ''})">Add source</button>
        </div>
        <div class="ui-field-grid">
          <div class="ui-field"><label>Attach to item IDs</label><input v-model="spells.attachItems" class="form-control form-control-sm" placeholder="800001, 800002"></div>
          <div class="ui-field">
            <label>Item field</label>
            <select v-model="spells.attachField" class="form-control form-control-sm">
              <option value="clickeffect">clickeffect</option>
              <option value="proceffect">proceffect</option>
              <option value="worneffect">worneffect</option>
              <option value="focuseffect">focuseffect</option>
              <option value="scrolleffect">scrolleffect</option>
            </select>
          </div>
        </div>
        <div class="ui-toolbar">
          <button type="button" class="btn btn-sm btn-dark" :disabled="busy" @click="runSpells(true)">Preview</button>
          <button type="button" class="btn btn-sm btn-primary" :disabled="busy" @click="runSpells(false)">Write spells</button>
        </div>
      </eq-window>

      <eq-window v-else-if="tab === 'npc'" title="NPC spell-set factory">
        <p class="zc-copy">
          Clone an <code>npc_spells</code> list and its entries. Spell IDs stay at or below 65535 so they fit in the entry column.
        </p>
        <div class="ui-field-grid">
          <div class="ui-field"><label>Search spell sets</label><input v-model="search.npc_spells" class="form-control form-control-sm" placeholder="Default Magician" @keyup.enter="doSearch('npc_spells')"></div>
          <div class="ui-field"><label>&nbsp;</label><button type="button" class="btn btn-sm btn-dark" :disabled="busy" @click="doSearch('npc_spells')">Search</button></div>
          <div class="ui-field"><label>Source list ID</label><input v-model.number="npcSpells.sourceId" type="number" class="form-control form-control-sm"></div>
          <div class="ui-field"><label>Prefix</label><input v-model="npcSpells.prefix" class="form-control form-control-sm" placeholder="T1"></div>
        </div>
        <div class="zc-missing" v-if="(results.npc_spells || []).length">
          <button v-for="row in results.npc_spells" :key="row.id" type="button" class="zc-pill ui-plain" @click="npcSpells.sourceId = row.id">
            {{ row.id }} {{ row.name }}
          </button>
        </div>
        <div class="zc-checks">
          <label><input type="checkbox" v-model="npcSpells.cloneSpells"> Clone each spell into the ≤65535 band</label>
        </div>
        <div class="ui-field-grid">
          <div class="ui-field"><label>Attach NPC IDs</label><textarea v-model="npcSpells.npcIds" class="form-control form-control-sm" rows="2"></textarea></div>
          <div class="ui-field"><label>Or zone IDs</label><textarea v-model="npcSpells.zoneIds" class="form-control form-control-sm" rows="2" placeholder="17"></textarea></div>
        </div>
        <div class="ui-toolbar">
          <button type="button" class="btn btn-sm btn-dark" :disabled="busy" @click="runNpcSpells(true)">Preview</button>
          <button type="button" class="btn btn-sm btn-primary" :disabled="busy" @click="runNpcSpells(false)">Write spell set</button>
        </div>
      </eq-window>

      <eq-window v-else-if="tab === 'kit'" title="Item kit studio">
        <p class="zc-copy">
          Each cell is a source item ID for that class and slot. Import fills from a live zone's loot.
        </p>
        <div class="ui-field-grid">
          <div class="ui-field"><label>Prefix</label><input v-model="kit.prefix" class="form-control form-control-sm" placeholder="T1"></div>
          <div class="ui-field"><label>First new item ID (0 = next free)</label><input v-model.number="kit.startId" type="number" class="form-control form-control-sm"></div>
          <div class="ui-field"><label>Copies</label><input v-model.number="kit.count" type="number" class="form-control form-control-sm"></div>
          <div class="ui-field"><label>Step multiplier</label><input v-model.number="kit.step" type="number" step="0.05" class="form-control form-control-sm"></div>
          <div class="ui-field"><label>Recipe id</label><input v-model="kit.recipeId" class="form-control form-control-sm" placeholder="classic-t1"></div>
          <div class="ui-field"><label>Recipe name</label><input v-model="kit.recipeName" class="form-control form-control-sm" placeholder="Classic T1"></div>
          <div class="ui-field"><label>Import zone</label><input v-model="kit.importZone" class="form-control form-control-sm" placeholder="17 or blackburrow"></div>
          <div class="ui-field"><label>Fill value</label><input v-model.number="kit.fillId" type="number" class="form-control form-control-sm"></div>
        </div>
        <div class="ui-field-grid">
          <div class="ui-field">
            <label>Search items</label>
            <input v-model="search.items" class="form-control form-control-sm" placeholder="Crown of King Tranix" @keyup.enter="doSearch('items')">
          </div>
          <div class="ui-field">
            <label>&nbsp;</label>
            <button type="button" class="btn btn-sm btn-dark" :disabled="busy" @click="doSearch('items')">Search</button>
          </div>
        </div>
        <div class="zc-missing" v-if="(results.items || []).length">
          <button v-for="row in results.items" :key="row.id" type="button" class="zc-pill ui-plain" @click="kit.fillId = row.id">
            {{ row.id }} {{ row.name }}
          </button>
        </div>
        <div class="zc-hint">Filled cells: <b>{{ filledCells.length }}</b></div>
        <div class="cov-wrap">
          <table class="eq-table cov-table kit-table">
            <thead>
            <tr>
              <th></th>
              <th v-for="cls in classes" :key="cls.name">
                {{ cls.name }}
                <button type="button" class="kit-mini ui-plain" @click="fillClass(cls.name)">fill</button>
              </th>
            </tr>
            </thead>
            <tbody>
            <tr v-for="slot in slots" :key="slot.name">
              <td>
                <b>{{ slot.name }}</b>
                <button type="button" class="kit-mini ui-plain" @click="fillSlot(slot.name)">fill</button>
              </td>
              <td v-for="cls in classes" :key="slot.name + cls.name" :class="cellId(slot.name, cls.name) ? 'cov-yes' : 'cov-no'">
                <input v-model.number="grid[cellKey(slot.name, cls.name)]" type="number" class="form-control form-control-sm kit-in">
              </td>
            </tr>
            </tbody>
          </table>
        </div>
        <div class="ui-toolbar">
          <button type="button" class="btn btn-sm btn-dark" :disabled="busy" @click="importKitZone">Import from zone</button>
          <button type="button" class="btn btn-sm btn-dark" @click="clearGrid">Clear grid</button>
          <button type="button" class="btn btn-sm btn-dark" :disabled="busy" @click="runItems(true)">Preview</button>
          <button type="button" class="btn btn-sm btn-primary" :disabled="busy" @click="runItems(false)">Write items</button>
          <router-link v-if="plan && plan.recipeId" class="btn btn-sm btn-dark" :to="ROUTE.ZONE_CONTROLLER_FACTORY">Open recipe in Tier Factory</router-link>
        </div>
      </eq-window>

      <eq-window v-else-if="tab === 'loot'" title="Loot composer">
        <p class="zc-copy">Trash and named tables attach separately. Leave named empty if you only want trash drops.</p>
        <div class="ui-field-grid">
          <div class="ui-field"><label>Trash table name</label><input v-model="loot.name" class="form-control form-control-sm" placeholder="classic-t1-trash"></div>
          <div class="ui-field"><label>Min cash</label><input v-model.number="loot.mincash" type="number" class="form-control form-control-sm"></div>
          <div class="ui-field"><label>Max cash</label><input v-model.number="loot.maxcash" type="number" class="form-control form-control-sm"></div>
          <div class="ui-field"><label>Drop probability</label><input v-model.number="loot.probability" type="number" class="form-control form-control-sm"></div>
          <div class="ui-field"><label>First loottable ID</label><input v-model.number="loot.startTableId" type="number" class="form-control form-control-sm"></div>
          <div class="ui-field"><label>First lootdrop ID</label><input v-model.number="loot.startDropId" type="number" class="form-control form-control-sm"></div>
        </div>
        <div class="ui-toolbar">
          <button type="button" class="btn btn-sm btn-dark" :disabled="!lastItemIds.length" @click="useLastKit">Use last kit items</button>
          <button type="button" class="btn btn-sm btn-dark" @click="loot.items.push({itemId: 0, chance: 100})">Add item</button>
        </div>
        <table class="eq-table mt-3" style="width: 100%">
          <thead><tr><th>Item ID</th><th>Chance</th><th></th></tr></thead>
          <tbody>
          <tr v-for="(it, i) in loot.items" :key="i">
            <td><input v-model.number="it.itemId" type="number" class="form-control form-control-sm"></td>
            <td><input v-model.number="it.chance" type="number" class="form-control form-control-sm"></td>
            <td><button type="button" class="btn btn-sm btn-outline-danger" @click="loot.items.splice(i, 1)">Remove</button></td>
          </tr>
          </tbody>
        </table>
        <h4 class="mt-3">Named loot</h4>
        <div class="ui-field-grid">
          <div class="ui-field"><label>Named table name</label><input v-model="namedLoot.name" class="form-control form-control-sm" placeholder="classic-t1-named"></div>
        </div>
        <div class="ui-toolbar">
          <button type="button" class="btn btn-sm btn-dark" @click="namedLoot.items.push({itemId: 0, chance: 100})">Add named item</button>
        </div>
        <table class="eq-table mt-3" style="width: 100%">
          <thead><tr><th>Item ID</th><th>Chance</th><th></th></tr></thead>
          <tbody>
          <tr v-for="(it, i) in namedLoot.items" :key="'n' + i">
            <td><input v-model.number="it.itemId" type="number" class="form-control form-control-sm"></td>
            <td><input v-model.number="it.chance" type="number" class="form-control form-control-sm"></td>
            <td><button type="button" class="btn btn-sm btn-outline-danger" @click="namedLoot.items.splice(i, 1)">Remove</button></td>
          </tr>
          </tbody>
        </table>
        <div class="ui-field-grid">
          <div class="ui-field"><label>Attach NPC IDs</label><textarea v-model="loot.npcIds" class="form-control form-control-sm" rows="2" placeholder="123, 456"></textarea></div>
          <div class="ui-field"><label>Or zone IDs</label><textarea v-model="loot.zoneIds" class="form-control form-control-sm" rows="2" placeholder="17, 31"></textarea></div>
          <div class="ui-field"><label>Recipe id</label><input v-model="loot.recipeId" class="form-control form-control-sm" placeholder="classic-t1"></div>
        </div>
        <div class="ui-toolbar">
          <button type="button" class="btn btn-sm btn-dark" :disabled="busy" @click="runLoot(true)">Preview</button>
          <button type="button" class="btn btn-sm btn-primary" :disabled="busy" @click="runLoot(false)">Write loot</button>
        </div>
      </eq-window>

      <eq-window v-else-if="tab === 'attach'" title="Bulk attach">
        <p class="zc-copy">Stock a merchant, or set loottable_id / npc_spells_id on the current target (trash by default). Preview lists every NPC that will change.</p>
        <div class="ui-field-grid">
          <div class="ui-field"><label>Merchant ID</label><input v-model.number="attach.merchantId" type="number" class="form-control form-control-sm" @change="loadMerchant"></div>
          <div class="ui-field"><label>Item IDs</label><textarea v-model="attach.itemIds" class="form-control form-control-sm" rows="2" placeholder="800001, 800002"></textarea></div>
        </div>
        <div class="zc-checks">
          <label><input type="checkbox" v-model="attach.replace"> Replace existing merchant stock</label>
          <label><input type="checkbox" v-model="attach.reload"> Reload / repop after write</label>
          <label><input type="checkbox" v-model="attach.syncZoneJson"> Sync zone item JSON spellid from PEQ</label>
        </div>
        <div class="ui-toolbar">
          <button type="button" class="btn btn-sm btn-dark" :disabled="!lastItemIds.length" @click="attach.itemIds = lastItemIds.join(', ')">Use last kit items</button>
          <button type="button" class="btn btn-sm btn-dark" :disabled="busy || !attach.merchantId" @click="loadMerchant">Refresh merchant</button>
        </div>
        <div class="zc-hint" v-if="merchant.ok">
          {{ merchant.npcName || "Merchant" }} {{ merchant.merchantId }} · {{ merchant.count }} in stock · next slot {{ merchant.nextSlot }}.
          {{ merchant.replaceNote }}
        </div>
        <table v-if="(merchant.stock || []).length" class="eq-table mt-3" style="width: 100%">
          <thead><tr><th>Slot</th><th>Item</th><th>Price</th><th>Chance</th></tr></thead>
          <tbody>
          <tr v-for="row in merchant.stock" :key="row.slot">
            <td class="tabular">{{ row.slot }}</td>
            <td>{{ row.name }} <span class="text-muted tabular">{{ row.itemId }}</span></td>
            <td class="tabular">{{ row.price }}</td>
            <td class="tabular">{{ row.probability }}</td>
          </tr>
          </tbody>
        </table>
        <div class="ui-field-grid">
          <div class="ui-field"><label>NPC IDs</label><textarea v-model="attach.npcIds" class="form-control form-control-sm" rows="2"></textarea></div>
          <div class="ui-field"><label>Zone IDs</label><textarea v-model="attach.zoneIds" class="form-control form-control-sm" rows="2"></textarea></div>
          <div class="ui-field"><label>Loottable ID</label><input v-model.number="attach.loottableId" type="number" class="form-control form-control-sm"></div>
          <div class="ui-field"><label>NPC spells ID</label><input v-model.number="attach.npcSpellsId" type="number" class="form-control form-control-sm"></div>
        </div>
        <div class="ui-toolbar">
          <button type="button" class="btn btn-sm btn-dark" :disabled="!lastTableId" @click="attach.loottableId = lastTableId">Use last loot table</button>
          <button type="button" class="btn btn-sm btn-dark" :disabled="busy" @click="runAttach(true)">Preview</button>
          <button type="button" class="btn btn-sm btn-primary" :disabled="busy" @click="runAttach(false)">Write attach</button>
        </div>
      </eq-window>

      <eq-window v-else-if="tab === 'pipeline'" title="One pipeline run">
        <p class="zc-copy">
          Spells → kit → NPC clones → trash/named loot → dual spell sets → merchant attach → export → reload → probe.
        </p>
        <div class="ui-field-grid">
          <div class="ui-field"><label>Prefix</label><input v-model="pipeline.prefix" class="form-control form-control-sm" placeholder="T1"></div>
          <div class="ui-field"><label>Step</label><input v-model.number="pipeline.step" type="number" step="0.05" class="form-control form-control-sm"></div>
          <div class="ui-field"><label>Recipe id</label><input v-model="pipeline.recipeId" class="form-control form-control-sm" placeholder="classic-t1"></div>
          <div class="ui-field"><label>Zone IDs</label><input v-model="pipeline.zoneIds" class="form-control form-control-sm" placeholder="17"></div>
          <div class="ui-field"><label>Merchant ID</label><input v-model.number="pipeline.merchantId" type="number" class="form-control form-control-sm"></div>
          <div class="ui-field"><label>Trash spell set source</label><input v-model.number="pipeline.npcSpellsId" type="number" class="form-control form-control-sm"></div>
          <div class="ui-field"><label>Named spell set source</label><input v-model.number="pipeline.npcSpellsNamedId" type="number" class="form-control form-control-sm"></div>
        </div>
        <div class="zc-checks">
          <label><input type="checkbox" v-model="pipeline.useSpells"> Include current spell sources</label>
          <label><input type="checkbox" v-model="pipeline.useKit"> Include filled kit cells</label>
          <label><input type="checkbox" v-model="pipeline.useLoot"> Build trash loot from kit</label>
          <label><input type="checkbox" v-model="pipeline.useNamedLoot"> Build named loot</label>
          <label><input type="checkbox" v-model="pipeline.cloneNpcs"> Clone NPCs (800000+)</label>
          <label><input type="checkbox" v-model="pipeline.retargetSpawns"> Retarget spawnentry to clones</label>
          <label><input type="checkbox" v-model="pipeline.replaceMerchant"> Replace merchant stock</label>
          <label><input type="checkbox" v-model="pipeline.exportClient"> Export spells_us.txt</label>
          <label><input type="checkbox" v-model="pipeline.reload"> Reload / repop</label>
          <label><input type="checkbox" v-model="pipeline.syncZoneJson"> Sync zone item JSON</label>
          <label><input type="checkbox" v-model="pipeline.probe"> Probe after write</label>
        </div>
        <div class="ui-toolbar">
          <button type="button" class="btn btn-sm btn-dark" :disabled="busy" @click="runPipeline(true)">Preview pipeline</button>
          <button type="button" class="btn btn-sm btn-primary" :disabled="busy" @click="runPipeline(false)">Write pipeline</button>
        </div>
      </eq-window>

      <eq-window v-else-if="tab === 'editor'" title="Spell-set editor">
        <p class="zc-copy">Load an npc_spells list, edit entries, then save as a new 800000+ list or replace the source.</p>
        <div class="ui-field-grid">
          <div class="ui-field"><label>List ID</label><input v-model.number="editor.id" type="number" class="form-control form-control-sm"></div>
          <div class="ui-field"><label>Name</label><input v-model="editor.name" class="form-control form-control-sm"></div>
          <div class="ui-field"><label>Prefix</label><input v-model="editor.prefix" class="form-control form-control-sm" placeholder="T1"></div>
          <div class="ui-field">
            <label>Search spells</label>
            <input v-model="search.spells" class="form-control form-control-sm" placeholder="Complete Heal" @keyup.enter="doSearch('spells')">
          </div>
        </div>
        <div class="ui-toolbar">
          <button type="button" class="btn btn-sm btn-dark" :disabled="busy || !editor.id" @click="loadSpellSet">Load</button>
          <button type="button" class="btn btn-sm btn-dark" @click="editor.entries.push({spellId: 0, minLevel: 1, maxLevel: 255, type: 1, manacost: -1, recastDelay: 0, priority: 0})">Add entry</button>
        </div>
        <div class="zc-missing" v-if="(results.spells || []).length">
          <button v-for="row in results.spells" :key="'ed' + row.id" type="button" class="zc-pill ui-plain" @click="editor.entries.push({spellId: row.id, name: row.name, minLevel: 1, maxLevel: 255, type: 1, manacost: -1, recastDelay: 0, priority: 0})">
            {{ row.id }} {{ row.name }}
          </button>
        </div>
        <table class="eq-table mt-3" style="width: 100%">
          <thead><tr><th>Spell</th><th>Min</th><th>Max</th><th>Type</th><th>Mana</th><th>Recast</th><th>Pri</th><th></th></tr></thead>
          <tbody>
          <tr v-for="(e, i) in editor.entries" :key="i">
            <td><input v-model.number="e.spellId" type="number" class="form-control form-control-sm" :title="e.name"></td>
            <td><input v-model.number="e.minLevel" type="number" class="form-control form-control-sm"></td>
            <td><input v-model.number="e.maxLevel" type="number" class="form-control form-control-sm"></td>
            <td><input v-model.number="e.type" type="number" class="form-control form-control-sm"></td>
            <td><input v-model.number="e.manacost" type="number" class="form-control form-control-sm"></td>
            <td><input v-model.number="e.recastDelay" type="number" class="form-control form-control-sm"></td>
            <td><input v-model.number="e.priority" type="number" class="form-control form-control-sm"></td>
            <td><button type="button" class="btn btn-sm btn-outline-danger" @click="editor.entries.splice(i, 1)">Remove</button></td>
          </tr>
          </tbody>
        </table>
        <div class="ui-toolbar">
          <button type="button" class="btn btn-sm btn-dark" :disabled="busy" @click="saveSpellSet(true, true)">Preview as new</button>
          <button type="button" class="btn btn-sm btn-primary" :disabled="busy" @click="saveSpellSet(false, true)">Write as new</button>
          <button type="button" class="btn btn-sm btn-outline-danger" :disabled="busy || !editor.id" @click="saveSpellSet(false, false)">Replace source</button>
        </div>
      </eq-window>

      <eq-window v-else-if="tab === 'test'" title="Test pawn and give">
        <p class="zc-copy">Spawn a SPIRE_TEST_ NPC at the zone safe point, or put minted items on a character inventory.</p>
        <h4>Pawn</h4>
        <div class="ui-field-grid">
          <div class="ui-field"><label>Zone ID</label><input v-model.number="pawn.zoneId" type="number" class="form-control form-control-sm"></div>
          <div class="ui-field"><label>Source NPC ID (0 = first trash)</label><input v-model.number="pawn.sourceNpcId" type="number" class="form-control form-control-sm"></div>
          <div class="ui-field"><label>Loottable ID</label><input v-model.number="pawn.loottableId" type="number" class="form-control form-control-sm"></div>
          <div class="ui-field"><label>NPC spells ID</label><input v-model.number="pawn.npcSpellsId" type="number" class="form-control form-control-sm"></div>
        </div>
        <div class="ui-toolbar">
          <button type="button" class="btn btn-sm btn-dark" :disabled="busy" @click="runPawn(true)">Preview pawn</button>
          <button type="button" class="btn btn-sm btn-primary" :disabled="busy" @click="runPawn(false)">Write pawn</button>
        </div>
        <h4 class="mt-3">Give to character</h4>
        <div class="ui-field-grid">
          <div class="ui-field"><label>Search characters</label><input v-model="search.characters" class="form-control form-control-sm" @keyup.enter="doSearch('characters')"></div>
          <div class="ui-field"><label>&nbsp;</label><button type="button" class="btn btn-sm btn-dark" :disabled="busy" @click="doSearch('characters')">Search</button></div>
          <div class="ui-field"><label>Character ID</label><input v-model.number="give.characterId" type="number" class="form-control form-control-sm"></div>
          <div class="ui-field"><label>Exact name</label><input v-model="give.name" class="form-control form-control-sm"></div>
          <div class="ui-field"><label>Item IDs</label><input v-model="give.itemIds" class="form-control form-control-sm" placeholder="800001, 800002"></div>
        </div>
        <div class="zc-missing" v-if="(results.characters || []).length">
          <button v-for="row in results.characters" :key="'ch' + row.id" type="button" class="zc-pill ui-plain" @click="give.characterId = row.id; give.name = row.name">
            {{ row.id }} {{ row.name }}
          </button>
        </div>
        <div class="ui-toolbar">
          <button type="button" class="btn btn-sm btn-dark" :disabled="!lastItemIds.length" @click="give.itemIds = lastItemIds.join(', ')">Use last kit items</button>
          <button type="button" class="btn btn-sm btn-dark" :disabled="busy" @click="runGive(true)">Preview give</button>
          <button type="button" class="btn btn-sm btn-primary" :disabled="busy" @click="runGive(false)">Write give</button>
        </div>
        <h4 class="mt-3">Probe</h4>
        <div class="ui-toolbar">
          <button type="button" class="btn btn-sm btn-dark" :disabled="busy" @click="runProbe">Probe last plan / zone</button>
        </div>
        <div class="ui-toolbar">
          <button type="button" class="btn btn-sm btn-dark" :disabled="busy" @click="runNpcs(true)">Preview NPC clones</button>
          <button type="button" class="btn btn-sm btn-primary" :disabled="busy" @click="runNpcs(false)">Write NPC clones</button>
        </div>
      </eq-window>

      <eq-window v-else title="Undo last run">
        <p class="zc-copy">
          Each write journals under <code>quests/global/ultimatedata/_spire_runs</code>.
          Undo deletes created rows and restores previous NPC loot/spell IDs.
        </p>
        <div class="ui-toolbar">
          <button type="button" class="btn btn-sm btn-dark" :disabled="busy" @click="loadRuns">Refresh</button>
        </div>
        <table class="eq-table" style="width: 100%">
          <thead><tr><th>Run</th><th>Kind</th><th>When</th><th>Summary</th><th></th></tr></thead>
          <tbody>
          <tr v-for="run in runs" :key="run.id">
            <td class="tabular">{{ run.id }}</td>
            <td>{{ run.kind }}</td>
            <td>{{ run.at }}</td>
            <td>{{ run.summary }} <span v-if="run.undone" class="text-muted">(undone)</span></td>
            <td>
              <button type="button" class="btn btn-sm btn-outline-danger" :disabled="busy || run.undone" @click="undoRun(run.id)">Undo</button>
            </td>
          </tr>
          <tr v-if="!runs.length">
            <td colspan="5" class="text-muted">No runs yet.</td>
          </tr>
          </tbody>
        </table>
      </eq-window>

      <eq-window v-if="plan" title="Plan">
        <div class="zc-hint" v-if="(plan.warnings || []).length">
          <div v-for="(w, i) in plan.warnings" :key="i">{{ w }}</div>
        </div>
        <div class="zc-hint" v-if="plan.export">{{ plan.export.note }}</div>
        <div class="zc-hint" v-if="plan.reload">{{ plan.reload.note }}</div>
        <div class="zc-hint" v-if="plan.probe">{{ plan.probe.note }}</div>
        <table v-if="(plan.lint || []).length" class="eq-table" style="width: 100%">
          <thead><tr><th>Lint</th><th>Code</th><th>ID</th><th>Message</th></tr></thead>
          <tbody>
          <tr v-for="(row, i) in plan.lint" :key="'lint' + i" :class="row.level === 'error' ? 'row-skip' : ''">
            <td>{{ row.level }}</td>
            <td>{{ row.code }}</td>
            <td class="tabular">{{ row.id || "—" }}</td>
            <td>{{ row.message }}</td>
          </tr>
          </tbody>
        </table>
        <table v-if="plan.probe && (plan.probe.checks || []).length" class="eq-table mt-3" style="width: 100%">
          <thead><tr><th>Probe</th><th>OK</th><th>Note</th></tr></thead>
          <tbody>
          <tr v-for="(c, i) in plan.probe.checks" :key="'pr' + i">
            <td>{{ c.name }}</td>
            <td>{{ c.ok ? "yes" : "no" }}</td>
            <td>{{ c.note }}</td>
          </tr>
          </tbody>
        </table>
        <table v-if="(plan.impact || []).length" class="eq-table mt-3" style="width: 100%">
          <thead><tr><th></th><th>Who changes</th><th>Role</th><th>Field</th><th>From</th><th>To</th><th>Note</th></tr></thead>
          <tbody>
          <tr v-for="(row, i) in plan.impact" :key="'imp' + i" :class="{'row-skip': skipImpact[row.key || i]}">
            <td><input v-if="row.id" type="checkbox" :checked="!skipImpact[row.key || i]" @change="toggleImpact(row, i)"></td>
            <td>{{ row.kind }} {{ row.name || row.id }}</td>
            <td><span v-if="row.role" class="role-pill" :class="'role-' + row.role">{{ row.role }}</span></td>
            <td>{{ row.field }}</td>
            <td class="tabular">{{ row.from }}</td>
            <td class="tabular">{{ row.to }}</td>
            <td>{{ row.note }}</td>
          </tr>
          </tbody>
        </table>
        <table v-if="(plan.clones || []).length" class="eq-table mt-3" style="width: 100%">
          <thead><tr><th>Kind</th><th>Source</th><th>New ID</th><th>Name</th><th>Diff</th></tr></thead>
          <tbody>
          <tr v-for="(c, i) in plan.clones" :key="i">
            <td>{{ c.kind }}</td>
            <td class="tabular">{{ c.sourceId }}</td>
            <td class="tabular">{{ c.newId }}</td>
            <td>{{ c.name }}</td>
            <td>{{ diffText(c.diff) }}</td>
          </tr>
          </tbody>
        </table>
        <table v-if="(plan.tables || []).length" class="eq-table mt-3" style="width: 100%">
          <thead><tr><th>Loot table</th><th>Table ID</th><th>Drop ID</th><th>Items</th></tr></thead>
          <tbody>
          <tr v-for="(t, i) in plan.tables" :key="i">
            <td>{{ t.name }}</td>
            <td class="tabular">{{ t.tableId }}</td>
            <td class="tabular">{{ t.dropId }}</td>
            <td class="tabular">{{ t.itemCount }}</td>
          </tr>
          </tbody>
        </table>
        <table v-if="(plan.dbWrites || []).length" class="eq-table mt-3" style="width: 100%">
          <thead><tr><th>Table</th><th>Action</th><th>ID</th><th>Note</th></tr></thead>
          <tbody>
          <tr v-for="(w, i) in plan.dbWrites" :key="'w' + i">
            <td>{{ w.table }}</td>
            <td>{{ w.action }}</td>
            <td class="tabular">{{ w.id }}</td>
            <td>{{ w.note }}</td>
          </tr>
          </tbody>
        </table>
      </eq-window>
    </eq-window>
  </content-area>
</template>

<script>
import EqWindow from "../../components/eq-ui/EQWindow"
import ContentArea from "../../components/layout/ContentArea"
import {ROUTE} from "@/routes"
import {ContentFactoryApi} from "../../app/content-factory"
import {ZoneControllerApi} from "../../app/zone-controller"

const defaultSlots = [
  {name: "Charm", bits: 1}, {name: "Ear", bits: 18}, {name: "Head", bits: 4}, {name: "Face", bits: 8},
  {name: "Neck", bits: 32}, {name: "Shoulders", bits: 64}, {name: "Arms", bits: 128}, {name: "Back", bits: 256},
  {name: "Wrist", bits: 1536}, {name: "Range", bits: 2048}, {name: "Hands", bits: 4096}, {name: "Primary", bits: 8192},
  {name: "Secondary", bits: 16384}, {name: "Ring", bits: 98304}, {name: "Chest", bits: 131072}, {name: "Legs", bits: 262144},
  {name: "Feet", bits: 524288}, {name: "Waist", bits: 1048576},
]
const defaultClasses = [
  {name: "WAR", bits: 1}, {name: "CLR", bits: 2}, {name: "PAL", bits: 4}, {name: "RNG", bits: 8},
  {name: "SK", bits: 16}, {name: "DRU", bits: 32}, {name: "MNK", bits: 64}, {name: "BRD", bits: 128},
  {name: "ROG", bits: 256}, {name: "SHM", bits: 512}, {name: "NEC", bits: 1024}, {name: "WIZ", bits: 2048},
  {name: "MAG", bits: 4096}, {name: "ENC", bits: 8192}, {name: "BST", bits: 16384}, {name: "BER", bits: 32768},
]

function emptyGrid() {
  const out = {}
  defaultSlots.forEach((slot) => {
    defaultClasses.forEach((cls) => {
      out[slot.name + "|" + cls.name] = 0
    })
  })
  return out
}

export default {
  name: "ContentFactory",
  components: {ContentArea, EqWindow},
  data() {
    return {
      ROUTE,
      tab: "census",
      error: "",
      busy: false,
      plan: null,
      board: {rows: []},
      runs: [],
      slots: defaultSlots,
      classes: defaultClasses,
      zoneQuery: "17",
      census: {},
      catalogKind: "items",
      catalog: {rows: []},
      merchant: {},
      search: {spells: "", items: "", npc_spells: "", characters: "", editor: ""},
      results: {spells: [], items: [], npc_spells: [], characters: []},
      roleChoices: ["trash", "named", "raid", "ignore", "all"],
      target: {roles: ["trash"], skipMerchants: true, includeRing: false, nameContains: "", excludeNpcIds: "", onlyNpcIds: ""},
      skipImpact: {},
      draftNote: "not saved yet",
      draftTimer: 0,
      draftReady: false,
      grid: emptyGrid(),
      spells: {
        prefix: "T1",
        startId: 0,
        count: 1,
        step: 1.25,
        npcCastable: false,
        exportClient: false,
        sources: [],
        attachItems: "",
        attachField: "clickeffect",
        scale: {effects: true, max: true, mana: true, duration: true, recast: false},
      },
      npcSpells: {sourceId: 0, prefix: "T1", cloneSpells: true, npcIds: "", zoneIds: ""},
      kit: {prefix: "T1", startId: 0, count: 1, step: 1.35, recipeId: "classic-t1", recipeName: "Classic T1", fillId: 0, importZone: "17"},
      loot: {name: "classic-t1-trash", mincash: 0, maxcash: 0, probability: 100, startTableId: 0, startDropId: 0, items: [{itemId: 0, chance: 100}], npcIds: "", zoneIds: "", recipeId: "classic-t1"},
      namedLoot: {name: "classic-t1-named", items: [{itemId: 0, chance: 100}]},
      attach: {merchantId: 0, itemIds: "", replace: false, npcIds: "", zoneIds: "", loottableId: 0, npcSpellsId: 0, reload: true, syncZoneJson: true},
      pipeline: {
        prefix: "T1", step: 1.35, recipeId: "classic-t1", zoneIds: "17", merchantId: 0,
        npcSpellsId: 0, npcSpellsNamedId: 0, useSpells: false, useKit: true, useLoot: true, useNamedLoot: false,
        cloneNpcs: true, retargetSpawns: true, replaceMerchant: false,
        exportClient: false, reload: true, syncZoneJson: true, probe: true,
      },
      editor: {id: 0, name: "", prefix: "T1", entries: []},
      pawn: {zoneId: 17, sourceNpcId: 0, loottableId: 0, npcSpellsId: 0},
      give: {characterId: 0, name: "", itemIds: ""},
    }
  },
  computed: {
    filledCells() {
      const out = []
      this.slots.forEach((slot) => {
        this.classes.forEach((cls) => {
          const id = Number(this.grid[this.cellKey(slot.name, cls.name)] || 0)
          if (id > 0) {
            out.push({slot: slot.name, class: cls.name, sourceId: id, slots: slot.bits, classes: cls.bits})
          }
        })
      })
      return out
    },
    lastItemIds() {
      return ((this.plan && this.plan.clones) || []).filter((c) => c.kind === "item").map((c) => c.newId)
    },
    lastTableId() {
      const tables = (this.plan && this.plan.tables) || []
      return tables.length ? tables[0].tableId : 0
    },
    draftPayload() {
      return {
        tab: this.tab,
        zoneQuery: this.zoneQuery,
        target: this.target,
        spells: this.spells,
        npcSpells: this.npcSpells,
        kit: this.kit,
        grid: this.grid,
        loot: this.loot,
        namedLoot: this.namedLoot,
        attach: this.attach,
        pipeline: this.pipeline,
        editor: this.editor,
        pawn: this.pawn,
        give: this.give,
      }
    },
  },
  watch: {
    draftPayload: {
      deep: true,
      handler() {
        this.queueDraft()
      },
    },
  },
  async mounted() {
    try {
      const meta = await ContentFactoryApi.meta()
      if (meta && meta.slots && meta.slots.length) {
        this.slots = meta.slots
      }
      if (meta && meta.classes && meta.classes.length) {
        this.classes = meta.classes
      }
      await this.loadDraft()
      await this.loadIds()
      await this.loadRuns()
      await this.loadCensus()
    } catch (e) {
      this.error = this.errText(e)
    }
  },
  methods: {
    cellKey(slot, cls) {
      return slot + "|" + cls
    },
    cellId(slot, cls) {
      return Number(this.grid[this.cellKey(slot, cls)] || 0)
    },
    spellList(rows) {
      return (rows || []).map((s) => s.name || s.id).slice(0, 6).join(", ")
    },
    fillSlot(slot) {
      if (!this.kit.fillId) {
        return
      }
      this.classes.forEach((cls) => {
        this.$set(this.grid, this.cellKey(slot, cls.name), this.kit.fillId)
      })
    },
    fillClass(cls) {
      if (!this.kit.fillId) {
        return
      }
      this.slots.forEach((slot) => {
        this.$set(this.grid, this.cellKey(slot.name, cls), this.kit.fillId)
      })
    },
    clearGrid() {
      this.grid = emptyGrid()
    },
    addSpell(row) {
      this.spells.sources.push({id: row.id, name: row.name || ""})
    },
    parseIds(raw) {
      return String(raw || "").split(/[,\s;]+/).map((n) => parseInt(n, 10)).filter((n) => n > 0)
    },
    hasRole(role) {
      return (this.target.roles || []).indexOf(role) !== -1
    },
    toggleRole(role) {
      if (role === "all") {
        this.target.roles = this.hasRole("all") ? ["trash"] : ["all"]
        return
      }
      const next = (this.target.roles || []).filter((r) => r !== "all")
      const i = next.indexOf(role)
      if (i === -1) {
        next.push(role)
      } else {
        next.splice(i, 1)
      }
      this.target.roles = next.length ? next : ["trash"]
    },
    isExcluded(id) {
      return this.parseIds(this.target.excludeNpcIds).indexOf(id) !== -1
    },
    toggleExclude(id) {
      const ids = this.parseIds(this.target.excludeNpcIds)
      const i = ids.indexOf(id)
      if (i === -1) {
        ids.push(id)
      } else {
        ids.splice(i, 1)
      }
      this.target.excludeNpcIds = ids.join(", ")
    },
    toggleImpact(row, i) {
      const key = row.key || i
      this.$set(this.skipImpact, key, !this.skipImpact[key])
      if (row.kind === "npc" && row.id) {
        this.toggleExclude(row.id)
      }
    },
    useCensusOnly() {
      const ids = (this.census.npcs || []).filter((n) => !this.isExcluded(n.id)).map((n) => n.id)
      this.target.onlyNpcIds = ids.join(", ")
    },
    openNeighbor(n) {
      this.zoneQuery = String(n.id || n.short)
      this.loadCensus()
    },
    targetPayload() {
      return {
        roles: this.target.roles,
        skipMerchants: this.target.skipMerchants,
        includeRing: this.target.includeRing,
        nameContains: this.target.nameContains,
        excludeNpcIds: this.parseIds(this.target.excludeNpcIds),
        onlyNpcIds: this.parseIds(this.target.onlyNpcIds),
      }
    },
    diffText(diff) {
      return (diff || []).map((d) => d.field + " " + d.from + "→" + d.to).join(", ")
    },
    applyDraft(body) {
      if (!body || typeof body !== "object") {
        return
      }
      const keys = ["tab", "zoneQuery", "target", "spells", "npcSpells", "kit", "loot", "namedLoot", "attach", "pipeline", "editor", "pawn", "give"]
      keys.forEach((key) => {
        if (body[key] && typeof body[key] === "object" && !Array.isArray(body[key]) && this[key] && typeof this[key] === "object") {
          this[key] = Object.assign({}, this[key], body[key])
        } else if (body[key] !== undefined && key !== "grid") {
          this[key] = body[key]
        }
      })
      if (body.grid && typeof body.grid === "object") {
        this.grid = Object.assign(emptyGrid(), body.grid)
      }
    },
    async loadDraft() {
      try {
        const data = await ContentFactoryApi.draft()
        let body = data && data.body
        if (typeof body === "string") {
          body = JSON.parse(body || "{}")
        }
        this.applyDraft(body)
        this.draftNote = data && data.relPath ? "loaded " + data.relPath : "empty draft"
      } catch (e) {
        this.draftNote = "draft not available"
      } finally {
        this.draftReady = true
      }
    },
    queueDraft() {
      if (!this.draftReady) {
        return
      }
      window.clearTimeout(this.draftTimer)
      this.draftTimer = window.setTimeout(() => {
        this.saveDraft()
      }, 800)
    },
    async saveDraft() {
      try {
        const data = await ContentFactoryApi.saveDraft(this.draftPayload)
        this.draftNote = data && data.ok ? "saved" : (data && data.error) || "save failed"
      } catch (e) {
        this.draftNote = this.errText(e)
      }
    },
    applyImportCells(cells) {
      (cells || []).forEach((cell) => {
        if (!cell.slot || !cell.class || !cell.sourceId) {
          return
        }
        this.$set(this.grid, this.cellKey(cell.slot, cell.class), cell.sourceId)
      })
    },
    async loadIds() {
      this.board = await ContentFactoryApi.ids()
    },
    async loadRuns() {
      const data = await ContentFactoryApi.runs()
      this.runs = data.runs || []
    },
    async loadCensus() {
      this.error = ""
      this.busy = true
      try {
        this.census = await ContentFactoryApi.census(this.zoneQuery || "17")
        if (this.census && this.census.error) {
          this.error = this.census.error
        } else if (this.census && this.census.zoneId) {
          this.pipeline.zoneIds = String(this.census.zoneId)
          this.kit.importZone = this.census.short || String(this.census.zoneId)
          this.pawn.zoneId = this.census.zoneId
        }
      } catch (e) {
        this.error = this.errText(e)
      } finally {
        this.busy = false
      }
    },
    async loadCatalog() {
      this.error = ""
      this.busy = true
      try {
        this.catalog = await ContentFactoryApi.catalog(this.catalogKind)
      } catch (e) {
        this.error = this.errText(e)
      } finally {
        this.busy = false
      }
    },
    async loadMerchant() {
      if (!this.attach.merchantId) {
        this.merchant = {}
        return
      }
      try {
        this.merchant = await ContentFactoryApi.merchant(this.attach.merchantId)
      } catch (e) {
        this.error = this.errText(e)
      }
    },
    async importKitZone() {
      this.error = ""
      this.busy = true
      try {
        const data = await ContentFactoryApi.importZone(this.kit.importZone || this.zoneQuery)
        if (data.error) {
          this.error = data.error
          return
        }
        this.applyImportCells(data.cells)
        this.tab = "kit"
      } catch (e) {
        this.error = this.errText(e)
      } finally {
        this.busy = false
      }
    },
    importCensusKit() {
      this.kit.importZone = this.zoneQuery
      this.importKitZone()
    },
    async doSearch(kind) {
      this.error = ""
      this.busy = true
      try {
        const data = await ContentFactoryApi.search(kind, this.search[kind] || "")
        this.$set(this.results, kind, data.rows || [])
      } catch (e) {
        this.error = this.errText(e)
      } finally {
        this.busy = false
      }
    },
    async runSpells(dry) {
      await this.run("spells", dry, () => ContentFactoryApi.spells({
        dryRun: dry,
        startId: this.spells.startId,
        prefix: this.spells.prefix,
        count: this.spells.count,
        step: this.spells.step,
        npcCastable: this.spells.npcCastable,
        exportClient: this.spells.exportClient,
        sources: this.spells.sources.filter((s) => s.id > 0),
        scale: this.spells.scale,
        attach: this.parseIds(this.spells.attachItems).map((itemId) => ({itemId, field: this.spells.attachField})),
      }))
    },
    async runNpcSpells(dry) {
      await this.run("npc_spells", dry, () => ContentFactoryApi.npcSpells({
        dryRun: dry,
        sourceId: this.npcSpells.sourceId,
        prefix: this.npcSpells.prefix,
        cloneSpells: this.npcSpells.cloneSpells,
        attachNpcIds: this.parseIds(this.npcSpells.npcIds),
        zoneIds: this.parseIds(this.npcSpells.zoneIds || this.pipeline.zoneIds),
        target: this.targetPayload(),
      }))
    },
    async runItems(dry) {
      await this.run("items", dry, () => ContentFactoryApi.items({
        dryRun: dry,
        startId: this.kit.startId,
        prefix: this.kit.prefix,
        count: this.kit.count,
        step: this.kit.step,
        recipeId: this.kit.recipeId,
        recipeName: this.kit.recipeName,
        cells: this.filledCells,
      }))
    },
    async runLoot(dry) {
      await this.run("loot", dry, () => ContentFactoryApi.loot({
        dryRun: dry,
        startTableId: this.loot.startTableId,
        startDropId: this.loot.startDropId,
        recipeId: this.loot.recipeId,
        attachNpcIds: this.parseIds(this.loot.npcIds),
        zoneIds: this.parseIds(this.loot.zoneIds || this.pipeline.zoneIds),
        target: this.targetPayload(),
        tables: [{
          name: this.loot.name,
          mincash: this.loot.mincash,
          maxcash: this.loot.maxcash,
          probability: this.loot.probability,
          target: "trash",
          items: this.loot.items.filter((it) => it.itemId > 0),
        }].concat((this.namedLoot.items || []).some((it) => it.itemId > 0) ? [{
          name: this.namedLoot.name,
          target: "named",
          probability: this.loot.probability,
          items: this.namedLoot.items.filter((it) => it.itemId > 0),
        }] : []),
      }))
    },
    async runAttach(dry) {
      await this.run("attach", dry, () => ContentFactoryApi.attach({
        dryRun: dry,
        merchantId: this.attach.merchantId,
        itemIds: this.parseIds(this.attach.itemIds),
        replace: this.attach.replace,
        npcIds: this.parseIds(this.attach.npcIds),
        zoneIds: this.parseIds(this.attach.zoneIds || this.pipeline.zoneIds),
        target: this.targetPayload(),
        loottableId: this.attach.loottableId,
        npcSpellsId: this.attach.npcSpellsId,
        reload: this.attach.reload,
        syncZoneJson: this.attach.syncZoneJson,
      }), this.attach.reload ? this.parseIds(this.attach.zoneIds || this.pipeline.zoneIds) : [])
    },
    async runPipeline(dry) {
      const zoneIds = this.parseIds(this.pipeline.zoneIds)
      await this.run("pipeline", dry, () => ContentFactoryApi.pipeline({
        dryRun: dry,
        prefix: this.pipeline.prefix,
        step: this.pipeline.step,
        recipeId: this.pipeline.recipeId,
        zoneIds,
        target: this.targetPayload(),
        merchantId: this.pipeline.merchantId,
        replaceMerchant: this.pipeline.replaceMerchant,
        exportClient: this.pipeline.exportClient,
        reload: this.pipeline.reload,
        syncZoneJson: this.pipeline.syncZoneJson,
        cloneNpcs: this.pipeline.cloneNpcs,
        retargetSpawns: this.pipeline.retargetSpawns,
        probe: this.pipeline.probe,
        spells: this.pipeline.useSpells ? {
          sources: this.spells.sources.filter((s) => s.id > 0),
          prefix: this.pipeline.prefix,
          step: this.pipeline.step,
          npcCastable: this.spells.npcCastable,
          scale: this.spells.scale,
        } : {sources: []},
        items: this.pipeline.useKit ? {
          prefix: this.pipeline.prefix,
          step: this.pipeline.step,
          recipeId: this.pipeline.recipeId,
          cells: this.filledCells,
        } : {cells: []},
        loot: this.pipeline.useLoot ? [{
          name: this.loot.name || (this.pipeline.prefix + "-trash"),
          mincash: this.loot.mincash,
          maxcash: this.loot.maxcash,
          probability: this.loot.probability,
          target: "trash",
          items: (this.loot.items.some((it) => it.itemId > 0) ? this.loot.items : this.lastItemIds.map((id) => ({itemId: id, chance: 100}))).filter((it) => it.itemId > 0),
        }] : [],
        namedLoot: this.pipeline.useNamedLoot && (this.namedLoot.items || []).some((it) => it.itemId > 0) ? [{
          name: this.namedLoot.name || (this.pipeline.prefix + "-named"),
          target: "named",
          probability: this.loot.probability,
          items: this.namedLoot.items.filter((it) => it.itemId > 0),
        }] : [],
        npcSpells: this.pipeline.npcSpellsId ? {
          sourceId: this.pipeline.npcSpellsId,
          prefix: this.pipeline.prefix,
          cloneSpells: this.npcSpells.cloneSpells,
          zoneIds,
          target: this.targetPayload(),
        } : {sourceId: 0},
        npcSpellsNamed: this.pipeline.npcSpellsNamedId ? {
          sourceId: this.pipeline.npcSpellsNamedId,
          prefix: this.pipeline.prefix,
          cloneSpells: this.npcSpells.cloneSpells,
          zoneIds,
        } : {sourceId: 0},
      }), this.pipeline.reload ? zoneIds : [])
    },
    async loadSpellSet() {
      this.error = ""
      this.busy = true
      try {
        const data = await ContentFactoryApi.spellSet(this.editor.id)
        if (data.error) {
          this.error = data.error
          return
        }
        this.editor.name = data.name || ""
        this.editor.entries = data.entries || []
        this.pipeline.npcSpellsId = this.editor.id
      } catch (e) {
        this.error = this.errText(e)
      } finally {
        this.busy = false
      }
    },
    async saveSpellSet(dry, asNew) {
      await this.run("spell_set", dry, () => ContentFactoryApi.saveSpellSet({
        dryRun: dry,
        id: this.editor.id,
        name: this.editor.name,
        prefix: this.editor.prefix,
        asNew,
        entries: this.editor.entries,
      }))
    },
    async runNpcs(dry) {
      await this.run("npcs", dry, () => ContentFactoryApi.npcs({
        dryRun: dry,
        prefix: this.pipeline.prefix,
        step: this.pipeline.step,
        zoneIds: this.parseIds(this.pipeline.zoneIds),
        target: this.targetPayload(),
        retargetSpawns: this.pipeline.retargetSpawns,
      }))
    },
    async runPawn(dry) {
      const zoneId = this.pawn.zoneId || this.parseIds(this.pipeline.zoneIds)[0] || 0
      await this.run("pawn", dry, () => ContentFactoryApi.pawn({
        dryRun: dry,
        zoneId,
        sourceNpcId: this.pawn.sourceNpcId,
        prefix: this.pipeline.prefix,
        loottableId: this.pawn.loottableId || this.lastTableId,
        npcSpellsId: this.pawn.npcSpellsId || this.pipeline.npcSpellsId,
      }), dry ? [] : [zoneId])
    },
    async runGive(dry) {
      await this.run("give", dry, () => ContentFactoryApi.give({
        dryRun: dry,
        characterId: this.give.characterId,
        name: this.give.name,
        itemIds: this.parseIds(this.give.itemIds),
      }))
    },
    async runProbe() {
      this.error = ""
      this.busy = true
      try {
        const clones = (this.plan && this.plan.clones) || []
        const data = await ContentFactoryApi.probe({
          zoneIds: this.parseIds(this.pipeline.zoneIds),
          spellIds: clones.filter((c) => c.kind === "spell").map((c) => c.newId),
          itemIds: clones.filter((c) => c.kind === "item").map((c) => c.newId),
          npcIds: clones.filter((c) => c.kind === "npc").map((c) => c.newId),
        })
        if (this.plan) {
          this.$set(this.plan, "probe", data)
        } else {
          this.plan = {ok: data.ok, kind: "probe", probe: data, clones: [], dbWrites: [], impact: []}
        }
      } catch (e) {
        this.error = this.errText(e)
      } finally {
        this.busy = false
      }
    },
    useLastKit() {
      this.loot.items = this.lastItemIds.map((id) => ({itemId: id, chance: 100}))
    },
    async undoRun(id) {
      if (!window.confirm("Undo run " + id + "? Created rows will be deleted.")) {
        return
      }
      await this.run("undo", false, () => ContentFactoryApi.undo(id))
      await this.loadRuns()
    },
    async run(kind, dry, fn, reloadZones) {
      this.error = ""
      this.busy = true
      try {
        this.skipImpact = {}
        this.plan = await fn()
        if (this.plan && this.plan.error) {
          this.error = this.plan.error
        }
        if (!dry && this.plan && this.plan.ok) {
          await this.loadIds()
          await this.loadRuns()
          if ((reloadZones || []).length) {
            await this.queueReload(reloadZones)
          }
        }
      } catch (e) {
        const data = e.response && e.response.data
        if (data && (data.kind || data.clones || data.dbWrites || data.impact)) {
          this.plan = data
        }
        this.error = this.errText(e)
      } finally {
        this.busy = false
      }
    },
    async queueReload(zoneIds) {
      try {
        const live = await ZoneControllerApi.apply({
          zoneIds,
          reloadQuests: true,
          queueCommands: ["refreshzonedata", "applyallchanges", "repop"],
          rebootEmpty: true,
        })
        if (this.plan) {
          this.$set(this.plan, "reload", {
            ok: !!live.ok,
            note: live.worldNote || "Reload queued.",
            reloaded: live.reloaded || [],
            warnings: live.warnings || [],
          })
        }
      } catch (e) {
        if (this.plan) {
          this.$set(this.plan, "reload", {ok: false, note: this.errText(e)})
        }
      }
    },
    errText(e) {
      const status = e.response && e.response.status
      if (status === 404) {
        return "Content Factory routes are not on the running backend yet. Restart Spire with start_spire_dev.bat."
      }
      return (e.response && e.response.data && e.response.data.error) || String(e)
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

.cov-wrap {
  overflow: auto;
  max-height: 640px;
}

.cov-table td, .cov-table th {
  text-align: center;
  min-width: 58px;
}

.cov-table td:first-child, .cov-table th:first-child {
  text-align: left;
  min-width: 90px;
}

.cov-yes {
  background: rgba(46, 160, 67, 0.18);
}

.cov-no {
  background: rgba(248, 81, 73, 0.12);
}

.kit-in {
  width: 72px;
  padding: 2px 4px;
}

.kit-mini {
  display: block;
  font-size: 10px;
  opacity: 0.7;
}

.mt-3 {
  margin-top: 1rem;
}

.tabular {
  font-variant-numeric: tabular-nums;
}

.row-skip {
  opacity: 0.45;
}

.role-pill {
  display: inline-block;
  padding: 0 6px;
  border-radius: 999px;
  font-size: 11px;
  text-transform: uppercase;
}

.role-trash { background: rgba(88, 166, 255, 0.18); }
.role-boss, .role-named { background: rgba(210, 153, 34, 0.22); }
.role-raid { background: rgba(248, 81, 73, 0.2); }
.role-ignore { background: rgba(110, 118, 129, 0.25); }
</style>
