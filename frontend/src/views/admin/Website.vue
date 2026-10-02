<template>
  <div>
    <eq-window title="Website">
      <p class="text-muted mb-3">
        The public site and Spire share the same peq database. Discord users, account links,
        character links, favorites, market, and achievements all live in <span class="mono">web_*</span>
        / <span class="mono">market_*</span> / <span class="mono">achievement_*</span> tables.
        This page is the admin side of that link — not a second copy of the data.
      </p>
      <div class="small mb-3" v-if="status.quests_root || status.website_root">
        <div v-if="status.quests_root">
          Live quests snapshot:
          <span class="mono privacy-hide">{{ status.quests_root }}</span>
        </div>
        <div v-if="status.website_root">
          Website snapshot:
          <span class="mono privacy-hide">{{ status.website_root }}</span>
        </div>
      </div>
      <div class="alert alert-warning" v-if="status.ok && !status.linked">
        This local database does not have <span class="mono">web_users</span> / account-link tables yet.
        They come from the live server migrations. EQ items, NPCs, and inventory still work;
        website linking will light up once those tables are on this connection.
      </div>
      <div class="web-table-grid">
        <div class="web-table-chip" v-for="t in status.tables || []" :key="t.name">
          <b>{{ t.name }}</b>
          <span v-if="t.present">{{ t.count }}</span>
          <span class="text-muted" v-else>missing</span>
        </div>
      </div>
    </eq-window>

    <eq-window title="Link Discord user → EQ account" class="mt-3">
      <div class="form-row">
        <div class="form-group col-md-4">
          <label>Web user</label>
          <input class="form-control" v-model="userQuery" placeholder="username, discord id, or #id" @keyup.enter="loadUsers">
        </div>
        <div class="form-group col-md-2 d-flex align-items-end">
          <button class="btn btn-dark btn-block" type="button" @click="loadUsers">Find users</button>
        </div>
        <div class="form-group col-md-4">
          <label>EQ account</label>
          <input class="form-control" v-model="accountQuery" placeholder="account name or id" @keyup.enter="loadAccounts">
        </div>
        <div class="form-group col-md-2 d-flex align-items-end">
          <button class="btn btn-dark btn-block" type="button" @click="loadAccounts">Find accounts</button>
        </div>
      </div>
      <div class="row">
        <div class="col-md-6">
          <table class="eq-table eq-highlight-rows bordered" v-if="users.length">
            <thead><tr><th></th><th>User</th><th>Role</th></tr></thead>
            <tbody>
            <tr v-for="u in users" :key="u.id" @click="selectedUser = u" :class="{active: selectedUser && selectedUser.id === u.id}">
              <td>{{ u.id }}</td>
              <td>{{ u.display_name || u.username }} <span class="privacy-hide text-muted">{{ u.discord_id }}</span></td>
              <td @click.stop>
                <select class="form-control form-control-sm" :value="u.role" :disabled="busy" @change="assignRole(u, $event)">
                  <option v-for="role in roleOptions" :key="role" :value="role">{{ role }}</option>
                </select>
              </td>
            </tr>
            </tbody>
          </table>
        </div>
        <div class="col-md-6">
          <table class="eq-table eq-highlight-rows bordered" v-if="accounts.length">
            <thead><tr><th></th><th>Account</th></tr></thead>
            <tbody>
            <tr v-for="a in accounts" :key="a.id" @click="selectedAccount = a" :class="{active: selectedAccount && selectedAccount.id === a.id}">
              <td>{{ a.id }}</td>
              <td>{{ a.name }}</td>
            </tr>
            </tbody>
          </table>
        </div>
      </div>
      <button class="btn btn-dark mt-2" type="button" :disabled="!selectedUser || !selectedAccount || busy" @click="linkAccount">
        Link selected user to account
      </button>
      <span class="text-danger ml-2" v-if="error">{{ error }}</span>
      <span class="text-success ml-2" v-if="notice">{{ notice }}</span>
    </eq-window>

    <eq-window title="Account links" class="mt-3">
      <div class="d-flex mb-2">
        <input class="form-control form-control-sm mr-2" v-model="linkQuery" placeholder="filter links" @keyup.enter="loadLinks">
        <button class="btn btn-sm btn-dark" type="button" @click="loadLinks">Refresh</button>
      </div>
      <table class="eq-table eq-highlight-rows bordered" v-if="links.length">
        <thead>
        <tr><th>User</th><th>EQ account</th><th>Character</th><th>By</th><th></th></tr>
        </thead>
        <tbody>
        <tr v-for="row in links" :key="row.id">
          <td>{{ row.display_name || row.username }} <span class="privacy-hide text-muted">#{{ row.web_user_id }}</span></td>
          <td>{{ row.eq_account_name }} ({{ row.eq_account_id }})</td>
          <td>{{ row.linked_character_name }}</td>
          <td>{{ row.linked_by }}</td>
          <td class="text-right">
            <button class="btn btn-sm btn-dark" type="button" @click="unlinkAccount(row.id)">Unlink</button>
          </td>
        </tr>
        </tbody>
      </table>
      <div class="text-muted" v-else>No account links on this database.</div>
    </eq-window>

    <eq-window title="Character links" class="mt-3">
      <div class="form-row">
        <div class="form-group col-md-4">
          <label>Character</label>
          <input class="form-control" v-model="charQuery" placeholder="character name or id" @keyup.enter="loadCharacters">
        </div>
        <div class="form-group col-md-2 d-flex align-items-end">
          <button class="btn btn-dark btn-block" type="button" @click="loadCharacters">Find</button>
        </div>
        <div class="form-group col-md-6 d-flex align-items-end">
          <button class="btn btn-dark" type="button" :disabled="!selectedUser || !selectedCharacter || busy" @click="linkCharacter">
            Link selected user to character
          </button>
        </div>
      </div>
      <table class="eq-table eq-highlight-rows bordered mb-3" v-if="characters.length">
        <thead><tr><th></th><th>Name</th><th>Account</th><th>Level</th></tr></thead>
        <tbody>
        <tr v-for="ch in characters" :key="ch.id" @click="selectedCharacter = ch" :class="{active: selectedCharacter && selectedCharacter.id === ch.id}">
          <td>{{ ch.id }}</td>
          <td>{{ ch.name }}</td>
          <td>{{ ch.account_id }}</td>
          <td>{{ ch.level }}</td>
        </tr>
        </tbody>
      </table>
      <table class="eq-table eq-highlight-rows bordered" v-if="charLinks.length">
        <thead><tr><th>User</th><th>Character</th><th>Source</th><th></th></tr></thead>
        <tbody>
        <tr v-for="row in charLinks" :key="row.id">
          <td>{{ row.display_name || row.username }}</td>
          <td>{{ row.character_name }} ({{ row.character_id }})</td>
          <td>{{ row.source }}</td>
          <td class="text-right">
            <button class="btn btn-sm btn-dark" type="button" @click="unlinkCharacter(row.id)">Unlink</button>
          </td>
        </tr>
        </tbody>
      </table>
    </eq-window>
  </div>
</template>

<script>
import EqWindow from "@/components/eq-ui/EQWindow"
import {WebsiteApi} from "@/app/website"

export default {
  name: "Website",
  components: {EqWindow},
  data() {
    return {
      status: {},
      users: [],
      accounts: [],
      characters: [],
      links: [],
      charLinks: [],
      userQuery: "",
      accountQuery: "",
      charQuery: "",
      linkQuery: "",
      selectedUser: null,
      selectedAccount: null,
      selectedCharacter: null,
      busy: false,
      error: "",
      notice: "",
      roleOptions: ["user", "admin"],
    }
  },
  async mounted() {
    await this.refresh()
  },
  methods: {
    async refresh() {
      try {
        this.status = await WebsiteApi.status()
        await this.loadUsers()
        await this.loadAccounts()
        await this.loadCharacters()
        await this.loadLinks()
        await this.loadCharLinks()
      } catch (e) {
        this.error = (e && e.response && e.response.data && e.response.data.error) || "Could not load website tables"
      }
    },
    async assignRole(user, event) {
      const nextRole = event && event.target ? event.target.value : ""
      if (!user || !nextRole || nextRole === user.role) {
        return
      }
      this.busy = true
      this.error = ""
      this.notice = ""
      const previous = user.role
      this.$set(user, "role", nextRole)
      try {
        await WebsiteApi.updateUserRole(user.id, nextRole)
        this.notice = "Set " + (user.display_name || user.username) + " to " + nextRole
      } catch (e) {
        this.$set(user, "role", previous)
        if (event && event.target) {
          event.target.value = previous
        }
        this.error = (e && e.response && e.response.data && e.response.data.error) || "Role update failed"
      }
      this.busy = false
    },
    async loadUsers() {
      const r = await WebsiteApi.users(this.userQuery)
      this.users = r.rows || []
    },
    async loadAccounts() {
      const r = await WebsiteApi.eqAccounts(this.accountQuery)
      this.accounts = r.rows || []
    },
    async loadCharacters() {
      const r = await WebsiteApi.eqCharacters(this.charQuery)
      this.characters = r.rows || []
    },
    async loadLinks() {
      const r = await WebsiteApi.accountLinks(this.linkQuery)
      this.links = r.rows || []
    },
    async loadCharLinks() {
      const r = await WebsiteApi.characterLinks(this.linkQuery)
      this.charLinks = r.rows || []
    },
    async linkAccount() {
      this.busy = true
      this.error = ""
      this.notice = ""
      try {
        await WebsiteApi.createAccountLink(this.selectedUser.id, this.selectedAccount.id)
        this.notice = "Linked " + this.selectedUser.username + " to " + this.selectedAccount.name
        await this.loadLinks()
      } catch (e) {
        this.error = (e && e.response && e.response.data && e.response.data.error) || "Link failed"
      }
      this.busy = false
    },
    async unlinkAccount(id) {
      await WebsiteApi.deleteAccountLink(id)
      await this.loadLinks()
    },
    async linkCharacter() {
      this.busy = true
      this.error = ""
      this.notice = ""
      try {
        await WebsiteApi.createCharacterLink(this.selectedUser.id, this.selectedCharacter.id)
        this.notice = "Linked " + this.selectedUser.username + " to " + this.selectedCharacter.name
        await this.loadCharLinks()
      } catch (e) {
        this.error = (e && e.response && e.response.data && e.response.data.error) || "Link failed"
      }
      this.busy = false
    },
    async unlinkCharacter(id) {
      await WebsiteApi.deleteCharacterLink(id)
      await this.loadCharLinks()
    },
  },
}
</script>

<style scoped>
.web-table-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 8px;
}
.web-table-chip {
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 8px;
  padding: 8px 10px;
  display: flex;
  justify-content: space-between;
  gap: 8px;
  font-size: 13px;
}
.eq-table tr.active td {
  outline: 1px solid #c9a24a;
}
</style>
