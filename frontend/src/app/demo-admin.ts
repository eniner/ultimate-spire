export function demoServerConfig() {
  return {
    server: {
      world: {
        shortname: "DEMO",
        longname: "Ultimate Spire Demo",
        address: "demo.local",
        localaddress: "127.0.0.1",
        key: "DEMO-SERVER-KEY-NOT-REAL",
        telnet: {enabled: "true", ip: "0.0.0.0", port: "9000"},
        tcp: {ip: "0.0.0.0", port: "9001"},
        loginserver1: {host: "login.eqemulator.net", port: "5998", legacy: "0", account: "demo", password: "demo"},
        loginserver2: {host: "", port: "5998", legacy: "0", account: "", password: ""},
        loginserver3: {host: "", port: "5998", legacy: "0", account: "", password: ""},
        loginserver4: {host: "", port: "5998", legacy: "0", account: "", password: ""},
        loginserver5: {host: "", port: "5998", legacy: "0", account: "", password: ""},
      },
      zones: {
        defaultstatus: "0",
        ports: {low: "7000", high: "7400"},
      },
      ucs: {host: "demo.local", port: "7778"},
      chatserver: {host: "demo.local", port: "7778"},
      mailserver: {host: "demo.local", port: "7778"},
      database: {db: "peq", host: "127.0.0.1", port: "3306", username: "demo", password: "demo"},
      qsdatabase: {db: "", host: "", port: "", username: "", password: ""},
      content_database: {db: "", host: "", port: "", username: "", password: ""},
    },
    "web-admin": {
      discord: {crash_log_webhook: "https://discord.example/demo-webhook"},
      quests: {hotReload: true},
      launcher: {
        runSharedMemory: true,
        runLoginserver: false,
        runQueryServ: false,
        runUcs: true,
        isRunning: true,
        minZoneProcesses: 5,
        staticZones: "blackburrow,qeynos,freporte",
        updateOpcodesOnStart: false,
        deleteLogFilesOlderThanDays: 14,
      },
    },
  }
}

export function demoAdminBody(path: string, method: string, params: any, tables: any) {
  if (path.indexOf("/admin/serverconfig") !== -1) {
    if (method !== "get") {
      return {ok: true, dryRun: true, message: "Server configuration previewed (demo is read-only)."}
    }
    return demoServerConfig()
  }
  if (path.indexOf("/admin/launcherconfig") !== -1) {
    return {
      runSharedMemory: true,
      runLoginserver: false,
      runQueryServ: false,
      runUcs: true,
      isRunning: true,
      spireLauncherStart: true,
      minZoneProcesses: 5,
      staticZones: "blackburrow,qeynos,freporte",
      updateOpcodesOnStart: false,
      deleteLogFilesOlderThanDays: 14,
    }
  }
  if (path.indexOf("/eqemuserver/dashboard-stats") !== -1) {
    return {
      accounts: ((tables && tables.account) || []).length,
      characters: ((tables && tables.character_data) || []).length,
      guilds: ((tables && tables.guilds) || []).length,
      items: ((tables && tables.item) || []).length,
      npcs: ((tables && tables.npc_types) || []).length,
    }
  }
  if (path.indexOf("/eqemuserver/server-stats") !== -1) {
    return {
      server_name: "Ultimate Spire Demo",
      zone_count: 8,
      players_online: 3,
      uptime: "Worldserver Uptime | 2 Days, 4 Hours, 18 Minutes",
      main_process_stats: [
        {description: "Spire Launcher", name: "eqemu-server:launcher", pid: 1001, cpu: 0.4, memory: 1.2, elapsed: 180000, optional: false, count: 1},
        {description: "World Server", name: "world", pid: 1002, cpu: 2.1, memory: 8.4, elapsed: 180000, optional: false, count: 1},
        {description: "Zones", name: "zone", pid: 1003, cpu: 6.8, memory: 22.1, elapsed: 170000, optional: false, count: 8},
        {description: "UCS", name: "ucs", pid: 1004, cpu: 0.3, memory: 1.1, elapsed: 170000, optional: true, count: 1},
        {description: "Login Server", name: "loginserver", pid: 0, cpu: 0, memory: 0, elapsed: 0, optional: true, count: 0},
        {description: "Query Server", name: "queryserv", pid: 0, cpu: 0, memory: 0, elapsed: 0, optional: true, count: 0},
      ],
    }
  }
  if (path.indexOf("/eqemuserver/system-all") !== -1) {
    return [{
      hostname: "demo-host",
      cpu: 18,
      mem_percent: 42,
      disk: [{readBytes: 120000000, writeBytes: 48000000}],
      net: [{name: "all", bytesRecv: 2500000, bytesSent: 1100000}],
    }]
  }
  if (path.indexOf("/eqemuserver/get-lock-status") !== -1) {
    return {locked: false}
  }
  if (path.indexOf("/eqemuserver/toggle-server-lock") !== -1) {
    return {locked: true, message: "Server lock previewed (demo is read-only)."}
  }
  if (path.indexOf("/eqemuserver/client-list") !== -1) {
    return {
      data: [
        {character_id: 1, name: "DemoWarrior", level: 12, class: 1, race: 1, online: 1, client_version: 6, ip: 2130706433, server: {zone_name: "blackburrow", zone_id: 17, instance_id: 0}},
        {character_id: 2, name: "DemoCleric", level: 10, class: 2, race: 1, online: 1, client_version: 6, ip: 2130706434, server: {zone_name: "qeynos", zone_id: 2, instance_id: 0}},
        {character_id: 3, name: "DemoWizard", level: 20, class: 12, race: 3, online: 1, client_version: 5, ip: 2130706435, server: {zone_name: "freporte", zone_id: 10, instance_id: 0}},
      ],
    }
  }
  if (path.indexOf("/eqemuserver/zoneserver-list") !== -1) {
    return [
      {id: 1, zone_id: 17, zone_name: "blackburrow", zone_long_name: "Blackburrow", zone_os_pid: 5101, is_static_zone: true, client_port: 7000, number_players: 1, instance_id: 0, clients: [], zone_server_address: "127.0.0.1:7000", compile_version: "demo", pid: 5101, name: "zone", cpu: 3.2, memory: 180000000, elapsed: 120000},
      {id: 2, zone_id: 2, zone_name: "qeynos", zone_long_name: "South Qeynos", zone_os_pid: 5102, is_static_zone: true, client_port: 7001, number_players: 1, instance_id: 0, clients: [], zone_server_address: "127.0.0.1:7001", compile_version: "demo", pid: 5102, name: "zone", cpu: 2.4, memory: 160000000, elapsed: 118000},
      {id: 3, zone_id: 10, zone_name: "freporte", zone_long_name: "East Freeport", zone_os_pid: 5103, is_static_zone: true, client_port: 7002, number_players: 1, instance_id: 0, clients: [], zone_server_address: "127.0.0.1:7002", compile_version: "demo", pid: 5103, name: "zone", cpu: 1.8, memory: 150000000, elapsed: 110000},
      {id: 4, zone_id: 54, zone_name: "gfaydark", zone_long_name: "The Greater Faydark", zone_os_pid: 5104, is_static_zone: false, client_port: 7003, number_players: 0, instance_id: 0, clients: [], zone_server_address: "127.0.0.1:7003", compile_version: "demo", pid: 5104, name: "zone", cpu: 0.6, memory: 120000000, elapsed: 90000},
      {id: 5, zone_id: 0, zone_name: "sleeping", zone_long_name: "Sleeping Zone", zone_os_pid: 5105, is_static_zone: false, client_port: 7004, number_players: 0, instance_id: 0, clients: [], zone_server_address: "127.0.0.1:7004", compile_version: "demo", pid: 5105, name: "zone", cpu: 0.1, memory: 80000000, elapsed: 80000},
    ]
  }
  if (path.indexOf("/eqemuserver/reload-types") !== -1) {
    return {
      data: [
        {command: "rules", description: "Rules"},
        {command: "logs", description: "Log settings"},
        {command: "quests", description: "Quests"},
        {command: "spells", description: "Spells"},
        {command: "npc", description: "NPCs"},
        {command: "loot", description: "Loot"},
        {command: "merchants", description: "Merchants"},
        {command: "doors", description: "Doors"},
        {command: "objects", description: "Objects"},
        {command: "zone", description: "Zone data"},
      ],
    }
  }
  if (path.indexOf("/eqemuserver/reload/") !== -1) {
    return {data: {message: "Reloading " + lastSeg(path) + " (demo preview)"}}
  }
  if (path.indexOf("/eqemuserver/logs") !== -1 && path.indexOf("/eqemuserver/log/") === -1) {
    return [
      {path: "eqemu_debug_world.log", name: "eqemu_debug_world.log", size: 42000, modified_time: Math.floor(Date.now() / 1000) - 120},
      {path: "eqemu_debug_zone.log", name: "eqemu_debug_zone.log", size: 88000, modified_time: Math.floor(Date.now() / 1000) - 30},
      {path: "crash.log", name: "crash.log", size: 1200, modified_time: Math.floor(Date.now() / 1000) - 3600},
    ]
  }
  if (path.indexOf("/eqemuserver/log-search/") !== -1) {
    const q = lastSeg(path)
    return [{
      file: "eqemu_debug_world.log",
      lines: [{line_number: 1, line: "[World] Demo search hit for " + q}, {line_number: 3, line: "[World] DemoWarrior entered world"}],
    }]
  }
  if (path.indexOf("/eqemuserver/log/") !== -1) {
    return {
      contents: "[World] Demo log line 1\n[Zone] blackburrow booted\n[World] DemoWarrior entered world\n[Say] DemoWarrior: Hail\n",
      path: lastSeg(path),
      cursor: 4,
    }
  }
  if (path.indexOf("/eqemuserver/version") !== -1) {
    return {version: "22.60.0", server_version: "22.60.0-demo", compile_date: "2026-01-01"}
  }
  if (path.indexOf("/eqemuserver/update-type") !== -1) {
    return {updateType: "release"}
  }
  if (path.indexOf("/eqemuserver/get-build-info") !== -1) {
    return {build: "release", status: "idle", last_build: "never"}
  }
  if (path.indexOf("/eqemuserver/build/branches") !== -1) {
    return ["master", "next"]
  }
  if (path.indexOf("/eqemuserver/build/current-branch") !== -1) {
    return {branch: "master"}
  }
  if (path.indexOf("/eqemuserver/player-event-logs/etl-settings") !== -1) {
    return {etl_settings: [{event_id: 1, table_name: "player_event_logs", enabled: true, etl_enabled: false, discord_id: 0, retention: 30}]}
  }
  if (path.indexOf("/admin/website/status") !== -1) {
    return {
      ok: true,
      linked: true,
      quests_root: "quests",
      website_root: "website",
      tables: [
        {name: "web_users", present: true, count: 3},
        {name: "web_account_links", present: true, count: 3},
        {name: "web_character_links", present: true, count: 4},
        {name: "market_listings", present: true, count: 6},
        {name: "achievement_progress", present: true, count: 8},
      ],
    }
  }
  if (path.indexOf("/admin/website/users") !== -1) {
    if (method !== "get") {
      return {ok: true, dryRun: true, message: "Website user change previewed (demo is read-only)."}
    }
    return {
      rows: [
        {id: 1, username: "demo-admin", display_name: "Demo Admin", discord_id: "1001", role: "admin"},
        {id: 2, username: "demo-mod", display_name: "Demo Mod", discord_id: "1002", role: "moderator"},
        {id: 3, username: "demo-player", display_name: "Demo Player", discord_id: "1003", role: "user"},
      ],
    }
  }
  if (path.indexOf("/admin/website/account-links") !== -1) {
    if (method !== "get") {
      return {ok: true, dryRun: true, message: "Account link previewed (demo is read-only)."}
    }
    return {rows: [{id: 1, web_user_id: 1, eq_account_id: 1}, {id: 2, web_user_id: 2, eq_account_id: 2}]}
  }
  if (path.indexOf("/admin/website/character-links") !== -1) {
    if (method !== "get") {
      return {ok: true, dryRun: true, message: "Character link previewed (demo is read-only)."}
    }
    return {rows: [{id: 1, web_user_id: 1, character_id: 1}, {id: 2, web_user_id: 3, character_id: 2}]}
  }
  if (path.indexOf("/admin/website/eq-accounts") !== -1) {
    return {rows: ((tables && tables.account) || []).map((a: any) => ({id: a.id, name: a.name}))}
  }
  if (path.indexOf("/admin/website/eq-characters") !== -1) {
    return {rows: ((tables && tables.character_data) || []).map((c: any) => ({id: c.id, name: c.name}))}
  }
  if (path.indexOf("/admin/system/host") !== -1) {
    return {hostname: "demo-host", os: "demo", cpu_count: 8, mem_total: 16}
  }
  if (path.indexOf("/admin/system/cpu") !== -1) {
    return {load: 18, cores: [12, 20, 8, 16]}
  }
  if (path.indexOf("/eqemuserver/manual-backup") !== -1 || path.indexOf("/backup/mysql") !== -1) {
    return {ok: true, dryRun: true, message: "Backup previewed (demo is read-only)."}
  }
  if (path.indexOf("/eqemuserver/server/") !== -1) {
    return {ok: true, dryRun: true, message: "Server control previewed (demo is read-only)."}
  }
  return undefined
}

function lastSeg(path: string) {
  const parts = path.replace(/\/$/, "").split("/").filter(Boolean)
  return decodeURIComponent(parts[parts.length - 1] || "")
}
