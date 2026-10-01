<template>
  <div>
    <eq-window title="Utilities" class="mb-3 p-3">
      <p class="mb-2">
        PHP Utilities were purge / economy / orphaned-grid helpers.
        Spire already has safer versions of those jobs.
      </p>
      <router-link class="btn btn-sm btn-dark" to="/editors">All editors</router-link>
    </eq-window>
    <eq-window title="Shortcuts" class="p-3">
      <div class="peq-util-grid">
        <router-link class="peq-util-card" to="/admin/backups">
          <strong>Backups</strong>
          <p>Database dumps and restore files.</p>
        </router-link>
        <router-link class="peq-util-card" to="/admin/database-backup">
          <strong>Manual dump</strong>
          <p>Run a mysqldump from Spire admin.</p>
        </router-link>
        <router-link class="peq-util-card" to="/admin/server-reload">
          <strong>Reload</strong>
          <p>World / zone reload instead of PHP import helpers.</p>
        </router-link>
        <router-link class="peq-util-card" to="/admin/players-online">
          <strong>Players online</strong>
          <p>Who is on the server right now.</p>
        </router-link>
        <router-link class="peq-util-card" to="/editors/players">
          <strong>Characters</strong>
          <p>Player sheets including xtargets. Inventory opens from a row.</p>
        </router-link>
        <router-link class="peq-util-card" to="/editors/raid-roles">
          <strong>Raid roles</strong>
          <p>Assister and marker assignment on raid_members.</p>
        </router-link>
        <router-link class="peq-util-card" to="/admin/website">
          <strong>Website roles</strong>
          <p>Assign user or admin on web_users.</p>
        </router-link>
        <router-link class="peq-util-card" to="/editors/accounts">
          <strong>Accounts</strong>
          <p>PHP "purge empty accounts" starts as an account search.</p>
        </router-link>
        <router-link class="peq-util-card" to="/editors/tradeskills">
          <strong>Recipe activity</strong>
          <p>Tradeskill recipe headers.</p>
        </router-link>
        <router-link class="peq-util-card" to="/editors/grids">
          <strong>Orphaned grids</strong>
          <p>Path grids by zone. Filter a zone and look for unused ids.</p>
        </router-link>
      </div>
    </eq-window>
  </div>
</template>

<script>
import EqWindow from "../../components/eq-ui/EQWindow"

export default {
  name: "PeqUtil",
  components: { EqWindow },
}
</script>

<style scoped>
.peq-util-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 10px;
}
.peq-util-card {
  display: block;
  padding: 12px;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(18, 22, 27, 0.45);
  color: inherit;
  text-decoration: none;
}
.peq-util-card:hover {
  border-color: rgba(201, 162, 74, 0.55);
  text-decoration: none;
  color: inherit;
}
.peq-util-card p {
  margin: 6px 0 0;
  font-size: 13px;
  opacity: 0.8;
}
</style>
