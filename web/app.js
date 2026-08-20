"use strict";

const $ = (id) => document.getElementById(id);

const state = {
  doc: null,
};

const fmt = (v, digits) => {
  if (v === null || v === undefined || Number.isNaN(Number(v))) return "—";
  return Number(v).toFixed(digits === undefined ? 4 : digits);
};

const num = (id) => {
  const raw = $(id).value.trim();
  if (raw === "") return null;
  const v = Number(raw);
  return Number.isFinite(v) ? v : null;
};

const setText = (cell, text) => {
  cell.textContent = text;
};

function fillForm(doc) {
  const w = doc.weather || {};
  $("netRadiation").value = w.netRadiation ?? "";
  $("soilHeatFlux").value = w.soilHeatFlux ?? 0;
  $("airTemperature").value = w.airTemperature ?? "";
  $("windSpeed").value = w.windSpeed ?? "";
  $("relativeHumidity").value = w.relativeHumidity ?? "";
  $("elevation").value = w.elevation ?? "";
  const growthDay = doc.crop && doc.crop.growthDay ? doc.crop.growthDay : "";
  $("growthDay").value = growthDay;
}

function buildPayload() {
  if (!state.doc) return null;
  const payload = JSON.parse(JSON.stringify(state.doc));
  if (payload.weather) {
    payload.weather.netRadiation = num("netRadiation") ?? payload.weather.netRadiation;
    payload.weather.soilHeatFlux = num("soilHeatFlux") ?? payload.weather.soilHeatFlux;
    payload.weather.airTemperature = num("airTemperature") ?? payload.weather.airTemperature;
    payload.weather.windSpeed = num("windSpeed") ?? payload.weather.windSpeed;
    payload.weather.relativeHumidity = num("relativeHumidity") ?? payload.weather.relativeHumidity;
    payload.weather.elevation = num("elevation") ?? payload.weather.elevation;
    if (!payload.weather.relativeHumidity) {
      delete payload.weather.relativeHumidity;
    }
    delete payload.weather.actualVaporPressure;
    delete payload.weather.dewpointTemperature;
  }
  if (payload.crop) {
    const day = num("growthDay");
    if (day !== null) {
      payload.crop.growthDay = day;
    }
  }
  delete payload.et0;
  return payload;
}

async function postJSON(path, body) {
  const res = await fetch(path, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  const data = await res.json().catch(() => ({}));
  return { ok: res.ok, status: res.status, data };
}

function renderError(message) {
  const box = $("form-error");
  box.textContent = message;
  box.classList.remove("hidden");
}

function clearError() {
  $("form-error").classList.add("hidden");
}

function renderET0(resp) {
  const tbody = $("et0-table").querySelector("tbody");
  tbody.innerHTML = "";
  const rows = [
    ["ET0", resp.result ? fmt(resp.result.et0, 4) : "—", "mm/d"],
    ["辐射项（0.408·Δ·(Rn−G)/分母）", resp.result ? fmt(resp.result.radiationTerm, 4) : "—", "mm/d"],
    ["空气动力项（γ·(900/(T+273))·u2·(es−ea)/分母）", resp.result ? fmt(resp.result.aerodynamicTerm, 4) : "—", "mm/d"],
    ["Δ（饱和水汽压斜率）", resp.result ? fmt(resp.result.delta, 5) : "—", "kPa/°C"],
    ["γ（湿度计常数）", resp.result ? fmt(resp.result.gamma, 5) : "—", "kPa/°C"],
    ["分母 Δ + γ·(1+0.34·u2)", resp.result ? fmt(resp.result.denominator, 5) : "—", "kPa/°C"],
    ["es / ea", resp.result ? `${fmt(resp.result.es, 4)} / ${fmt(resp.result.ea, 4)}` : "—", "kPa"],
    ["es − ea（饱和差）", resp.result ? fmt(resp.result.deficit, 4) : "—", "kPa"],
    ["Rn − G", resp.result ? fmt(resp.result.availableEnergy, 4) : "—", "MJ/(m²·d)"],
  ];
  for (const [name, value, unit] of rows) {
    const tr = document.createElement("tr");
    const tdName = document.createElement("td");
    tdName.textContent = name;
    const tdValue = document.createElement("td");
    tdValue.textContent = value;
    const tdUnit = document.createElement("td");
    tdUnit.textContent = unit;
    tr.append(tdName, tdValue, tdUnit);
    tbody.append(tr);
  }
}

function renderETc(resp) {
  const tbody = $("etc-table").querySelector("tbody");
  tbody.innerHTML = "";
  if (!resp.crop) {
    const tr = document.createElement("tr");
    const td = document.createElement("td");
    td.colSpan = 3;
    td.textContent = "文档未带 crop 块";
    tr.append(td);
    tbody.append(tr);
    $("stage-table").querySelector("tbody").innerHTML = "";
    return;
  }
  const c = resp.crop;
  const rows = [
    ["ET0", fmt(c.et0, 4), "mm/d"],
    ["Kc", fmt(c.kc, 3), ""],
    ["阶段", String(c.stage), ""],
    ["生长日", String(c.growthDay), "d"],
    ["Ks（水分胁迫）", fmt(c.stressCoefficient, 3), ""],
    ["ETc 无胁迫（Kc·ET0）", fmt(c.etcPotential, 4), "mm/d"],
    ["ETc", fmt(c.etc, 4), "mm/d"],
  ];
  for (const [name, value, unit] of rows) {
    const tr = document.createElement("tr");
    const tdName = document.createElement("td");
    tdName.textContent = name;
    const tdValue = document.createElement("td");
    tdValue.textContent = value;
    const tdUnit = document.createElement("td");
    tdUnit.textContent = unit;
    tr.append(tdName, tdValue, tdUnit);
    tbody.append(tr);
  }
  const stageBody = $("stage-table").querySelector("tbody");
  stageBody.innerHTML = "";
  for (const row of c.stageTable || []) {
    const tr = document.createElement("tr");
    const cells = [
      String(row.stage),
      `${row.firstDay}..${row.lastDay}`,
      fmt(row.kc, 3),
      fmt(row.etc, 4),
    ];
    for (const text of cells) {
      const td = document.createElement("td");
      td.textContent = text;
      tr.append(td);
    }
    stageBody.append(tr);
  }
}

function renderWindSweep(sweep) {
  const tbody = $("wind-table").querySelector("tbody");
  tbody.innerHTML = "";
  if (!sweep || sweep.length === 0) {
    const tr = document.createElement("tr");
    const td = document.createElement("td");
    td.colSpan = 5;
    td.textContent = "本次请求未返回风速扫描";
    tr.append(td);
    tbody.append(tr);
    return;
  }
  for (const point of sweep) {
    const tr = document.createElement("tr");
    const cells = [
      fmt(point.windSpeed, 2),
      fmt(point.et0, 4),
      fmt(point.radiationTerm, 4),
      fmt(point.aerodynamicTerm, 4),
      fmt(100 * point.aerodynamicShare, 1) + "%",
    ];
    for (const text of cells) {
      const td = document.createElement("td");
      td.textContent = text;
      tr.append(td);
    }
    tbody.append(tr);
  }
}

async function compute() {
  clearError();
  const payload = buildPayload();
  if (!payload) {
    renderError("请先加载算例");
    return;
  }
  try {
    const [et0Resp, etcResp] = await Promise.all([
      postJSON("/api/et0", payload),
      postJSON("/api/etc", payload),
    ]);
    if (!et0Resp.ok) {
      renderError("ET0 计算失败: " + (et0Resp.data.error || "HTTP " + et0Resp.status));
      return;
    }
    if (!etcResp.ok) {
      renderError("ETc 计算失败: " + (etcResp.data.error || "HTTP " + etcResp.status));
      return;
    }
    renderET0(et0Resp.data);
    renderETc(etcResp.data);
    renderWindSweep(et0Resp.data.windSweep);
  } catch (err) {
    renderError("请求失败: " + err.message);
  }
}

async function loadExample() {
  try {
    const res = await fetch("/api/example");
    if (!res.ok) {
      renderError("加载算例失败: HTTP " + res.status);
      return;
    }
    state.doc = await res.json();
    fillForm(state.doc);
    await compute();
  } catch (err) {
    renderError("加载算例失败: " + err.message);
  }
}

function wire() {
  $("load-example").addEventListener("click", loadExample);
  $("compute").addEventListener("click", compute);
  for (const id of [
    "netRadiation",
    "soilHeatFlux",
    "airTemperature",
    "windSpeed",
    "relativeHumidity",
    "elevation",
    "growthDay",
  ]) {
    $(id).addEventListener("change", compute);
    $(id).addEventListener("input", compute);
  }
}

wire();
loadExample();
