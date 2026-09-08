const API = 'http://localhost:8080';

function fmt(n) {
  if (n == null) return '—';
  return new Intl.NumberFormat('es-CO', { style: 'currency', currency: 'COP', maximumFractionDigits: 0 }).format(n);
}
function fmtDate(d) {
  if (!d) return '—';
  return new Date(d).toLocaleDateString('es-CO');
}
function badgeEstado(e) {
  if (!e) return '—';
  const key = e.toLowerCase().replace('aprobado_con_observaciones', 'observaciones');
  return `<span class="badge badge-${key}">${e.replace(/_/g, ' ')}</span>`;
}
function toast(msg, type = 'success') {
  const el = document.getElementById('appToast');
  const iconMap = { success: 'bi-check-circle-fill', error: 'bi-x-circle-fill', info: 'bi-info-circle-fill' };
  el.className = `toast ${type}`;
  document.getElementById('toastIcon').className = `toast-icon bi ${iconMap[type] || iconMap.info}`;
  document.getElementById('toastMsg').textContent = msg;
  bootstrap.Toast.getOrCreateInstance(el, { delay: 3500 }).show();
}
async function apiFetch(url, opts = {}) {
  const res = await fetch(API + url, { headers: { 'Content-Type': 'application/json' }, ...opts });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || `Error ${res.status}`);
  return data;
}
function openModal(id) { bootstrap.Modal.getOrCreateInstance(document.getElementById(id)).show(); }
function closeModal(id) { bootstrap.Modal.getOrCreateInstance(document.getElementById(id)).hide(); }
function loading() { return `<div class="loading-state"><div class="spinner"></div> Cargando...</div>`; }
function empty(msg = 'Sin registros') { return `<div class="empty-state">${msg}</div>`; }

// ─── NAVEGACION ───
document.querySelectorAll('.sidebar-link').forEach(link => {
  link.addEventListener('click', e => {
    e.preventDefault();
    document.querySelectorAll('.sidebar-link').forEach(l => l.classList.remove('active'));
    link.classList.add('active');
    document.querySelectorAll('[id^="section-"]').forEach(s => s.classList.add('d-none'));
    const sec = document.getElementById(`section-${link.dataset.section}`);
    sec.classList.remove('d-none');
    sec.classList.remove('animate-in');
    void sec.offsetWidth;
    sec.classList.add('animate-in');
    ({ dashboard: cargarDashboard, clientes: cargarClientes, vehiculos: cargarVehiculos,
       ordenes: cargarOrdenes, repuestos: cargarRepuestos, cda: cargarCDA })[link.dataset.section]?.();
  });
});

// ─── DASHBOARD ───
async function cargarDashboard() {
  try {
    const [clientes, vehiculos, ordenes, stock] = await Promise.all([
      apiFetch('/api/clientes'), apiFetch('/api/vehiculos'),
      apiFetch('/api/ordenes'), apiFetch('/api/repuestos/bajo-stock')
    ]);
    document.getElementById('stat-clientes').textContent = clientes.length;
    document.getElementById('stat-vehiculos').textContent = vehiculos.length;
    document.getElementById('stat-stock-bajo').textContent = stock.length;
    document.getElementById('stat-ordenes-activas').textContent =
      ordenes.filter(o => ['PENDIENTE','EN_PROCESO'].includes(o.estado)).length;

    document.getElementById('dashboard-ordenes').innerHTML = ordenes.length === 0
      ? empty('Sin ordenes recientes')
      : `<table class="data-table"><thead><tr><th>#</th><th>Cliente</th><th>Placa</th><th>Estado</th><th>Total</th></tr></thead><tbody>
          ${ordenes.slice(0,8).map(o => `<tr>
            <td style="color:var(--text-2)">${o.ordenId}</td>
            <td>${o.cliente ?? '—'}</td>
            <td><code style="color:var(--gold);font-size:0.78rem">${o.placa ?? '—'}</code></td>
            <td>${badgeEstado(o.estado)}</td>
            <td>${fmt(o.totalFacturado)}</td>
          </tr>`).join('')}
         </tbody></table>`;

    document.getElementById('dashboard-stock').innerHTML = stock.length === 0
      ? `<div class="empty-state" style="color:var(--success)">Todo el stock OK</div>`
      : `<table class="data-table"><thead><tr><th>Repuesto</th><th>Stock</th><th>Min</th></tr></thead><tbody>
          ${stock.map(r => `<tr>
            <td>${r.nombre ?? r.Nombre}</td>
            <td style="color:var(--danger);font-weight:600">${r.stock_actual ?? r.StockActual}</td>
            <td class="muted">${r.stock_minimo ?? r.StockMinimo}</td>
          </tr>`).join('')}
         </tbody></table>`;
  } catch(err) { toast(err.message, 'error'); }
}

// ─── CLIENTES ───
async function cargarClientes() {
  document.getElementById('tabla-clientes').innerHTML = loading();
  try {
    const data = await apiFetch('/api/clientes');
    if (!data.length) { document.getElementById('tabla-clientes').innerHTML = empty(); return; }
    document.getElementById('tabla-clientes').innerHTML =
      `<table class="data-table">
        <thead><tr><th>ID</th><th>Cedula</th><th>Nombre</th><th>Telefono</th><th>Email</th><th>Direccion</th></tr></thead>
        <tbody>${data.map(c => `<tr>
          <td class="muted">${c.clienteId}</td>
          <td>${c.cedula}</td>
          <td><strong>${c.nombre} ${c.apellido}</strong></td>
          <td class="muted">${c.telefono ?? '—'}</td>
          <td class="muted">${c.email ?? '—'}</td>
          <td class="muted">${c.direccion ?? '—'}</td>
        </tr>`).join('')}</tbody>
      </table>`;
  } catch(err) { toast(err.message, 'error'); }
}
async function guardarCliente() {
  try {
    await apiFetch('/api/clientes', { method: 'POST', body: JSON.stringify({
      cedula:    document.getElementById('c-cedula').value,
      nombre:    document.getElementById('c-nombre').value,
      apellido:  document.getElementById('c-apellido').value,
      telefono:  document.getElementById('c-telefono').value,
      email:     document.getElementById('c-email').value,
      direccion: document.getElementById('c-direccion').value,
      placa:     document.getElementById('c-placa').value,
      marca:     document.getElementById('c-marca').value,
      modelo:    document.getElementById('c-modelo').value,
      anio:      +document.getElementById('c-anio').value,
      color:     document.getElementById('c-color').value,
      cilindraje:+document.getElementById('c-cilindraje').value || null,
      kmActual:  +document.getElementById('c-km').value
    })});
    toast('Cliente registrado'); closeModal('modal-cliente'); cargarClientes();
  } catch(err) { toast(err.message, 'error'); }
}

// ─── VEHICULOS ───
async function cargarVehiculos() {
  document.getElementById('tabla-vehiculos').innerHTML = loading();
  try {
    const placa = document.getElementById('buscar-placa')?.value ?? '';
    const data = await apiFetch('/api/vehiculos' + (placa ? `?placa=${encodeURIComponent(placa)}` : ''));
    if (!data.length) { document.getElementById('tabla-vehiculos').innerHTML = empty(); return; }
    document.getElementById('tabla-vehiculos').innerHTML =
      `<table class="data-table">
        <thead><tr><th>Placa</th><th>Marca</th><th>Modelo</th><th>Año</th><th>Km</th><th>Cliente</th></tr></thead>
        <tbody>${data.map(v => `<tr>
          <td><code style="color:var(--gold);font-size:0.78rem">${v.placa}</code></td>
          <td>${v.marca}</td>
          <td>${v.modelo}</td>
          <td class="muted">${v.anio}</td>
          <td class="muted">${(v.kmActual || 0).toLocaleString()} km</td>
          <td class="muted">${v.nombreCliente ?? '—'}</td>
        </tr>`).join('')}</tbody>
      </table>`;
  } catch(err) { toast(err.message, 'error'); }
}
async function guardarVehiculo() {
  try {
    await apiFetch('/api/vehiculos', { method: 'POST', body: JSON.stringify({
      clienteId: +document.getElementById('v-clienteid').value,
      placa:     document.getElementById('v-placa').value,
      marca:     document.getElementById('v-marca').value,
      modelo:    document.getElementById('v-modelo').value,
      anio:      +document.getElementById('v-anio').value,
      color:     document.getElementById('v-color').value,
      kmActual:  +document.getElementById('v-km').value
    })});
    toast('Vehiculo registrado'); closeModal('modal-vehiculo'); cargarVehiculos();
  } catch(err) { toast(err.message, 'error'); }
}

// ─── ORDENES ───
async function cargarOrdenes() {
  document.getElementById('tabla-ordenes').innerHTML = loading();
  try {
    const estado = document.getElementById('filtro-estado')?.value ?? '';
    const data = await apiFetch('/api/ordenes' + (estado ? `?estado=${estado}` : ''));
    if (!data.length) { document.getElementById('tabla-ordenes').innerHTML = empty(); return; }
    document.getElementById('tabla-ordenes').innerHTML =
      `<table class="data-table">
        <thead><tr><th>#</th><th>Ingreso</th><th>Placa</th><th>Vehiculo</th><th>Cliente</th><th>Estado</th><th>Total</th><th></th></tr></thead>
        <tbody>${data.map(o => `<tr>
          <td class="muted">${o.ordenId}</td>
          <td class="muted">${fmtDate(o.fechaIngreso)}</td>
          <td><code style="color:var(--gold);font-size:0.78rem">${o.placa ?? '—'}</code></td>
          <td>${o.vehiculoDesc ?? '—'}</td>
          <td>${o.cliente ?? '—'}</td>
          <td>${badgeEstado(o.estado)}</td>
          <td>${fmt(o.totalFacturado)}</td>
          <td><div class="actions"><button class="btn btn-icon" onclick="verOrden(${o.ordenId})"><i class="bi bi-eye"></i></button></div></td>
        </tr>`).join('')}</tbody>
      </table>`;
  } catch(err) { toast(err.message, 'error'); }
}
async function guardarOrden() {
  try {
    await apiFetch('/api/ordenes', { method: 'POST', body: JSON.stringify({
      vehiculoId:   +document.getElementById('o-vehiculoid').value,
      empleadoId:   +document.getElementById('o-empleadoid').value,
      kmIngreso:    +document.getElementById('o-km').value,
      fechaEstimada: document.getElementById('o-fecha').value || null,
      diagnostico:   document.getElementById('o-diagnostico').value
    })});
    toast('Orden creada'); closeModal('modal-orden'); cargarOrdenes();
  } catch(err) { toast(err.message, 'error'); }
}

let _ordenActual = null;
async function verOrden(id) {
  _ordenActual = id;
  document.getElementById('detalle-orden-id').textContent = `#${id}`;
  document.getElementById('detalle-orden-body').innerHTML = loading();
  document.getElementById('btn-cerrar-orden').classList.add('d-none');
  openModal('modal-detalle-orden');
  try {
    const d = await apiFetch(`/api/ordenes/${id}`);
    const o = d.orden;
    const editable = ['PENDIENTE','EN_PROCESO'].includes(o.estado);
    document.getElementById('btn-cerrar-orden').classList.toggle('d-none', !editable);
    if (editable) { await Promise.all([cargarSelectServicios(), cargarSelectRepuestos()]); }

    document.getElementById('detalle-orden-body').innerHTML = `
      <div class="detail-meta">
        <div class="detail-meta-item"><span class="detail-meta-label">Estado</span><div class="detail-meta-value">${badgeEstado(o.estado)}</div></div>
        <div class="detail-meta-item"><span class="detail-meta-label">Cliente</span><div class="detail-meta-value">${o.cliente ?? '—'}</div></div>
        <div class="detail-meta-item"><span class="detail-meta-label">Placa</span><div class="detail-meta-value" style="color:var(--gold)">${o.placa ?? '—'}</div></div>
        <div class="detail-meta-item"><span class="detail-meta-label">Km Ingreso</span><div class="detail-meta-value">${(o.kmIngreso || 0).toLocaleString()} km</div></div>
        <div class="detail-meta-item"><span class="detail-meta-label">Fecha Ingreso</span><div class="detail-meta-value">${fmtDate(o.fechaIngreso)}</div></div>
        <div class="detail-meta-item"><span class="detail-meta-label">Responsable</span><div class="detail-meta-value">${o.empleado ?? '—'}</div></div>
      </div>
      ${o.diagnostico ? `<div class="sub-panel mb-4"><div class="sub-panel-header"><span class="sub-panel-title">Diagnostico</span></div><div style="padding:1rem 1.2rem;font-size:0.85rem;color:var(--text-2)">${o.diagnostico}</div></div>` : ''}
      <div style="display:grid;grid-template-columns:1fr 1fr;gap:1rem">
        <div class="sub-panel">
          <div class="sub-panel-header">
            <span class="sub-panel-title">Servicios</span>
            ${editable ? `<button class="btn btn-icon" style="color:var(--gold)" onclick="openModal('modal-add-servicio')"><i class="bi bi-plus"></i></button>` : ''}
          </div>
          <table class="data-table">
            <thead><tr><th>Servicio</th><th>Cant</th><th>Subtotal</th></tr></thead>
            <tbody>${(d.servicios || []).map(s => `<tr><td>${s.servicio}</td><td class="muted">${s.cantidad}</td><td>${fmt(s.subtotal)}</td></tr>`).join('') || `<tr><td colspan="3" class="muted" style="text-align:center;padding:1.5rem">Sin servicios</td></tr>`}</tbody>
          </table>
        </div>
        <div class="sub-panel">
          <div class="sub-panel-header">
            <span class="sub-panel-title">Repuestos</span>
            ${editable ? `<button class="btn btn-icon" style="color:var(--gold)" onclick="openModal('modal-add-repuesto')"><i class="bi bi-plus"></i></button>` : ''}
          </div>
          <table class="data-table">
            <thead><tr><th>Repuesto</th><th>Cant</th><th>Subtotal</th></tr></thead>
            <tbody>${(d.repuestos || []).map(r => `<tr><td>${r.repuesto}</td><td class="muted">${r.cantidad}</td><td>${fmt(r.subtotal)}</td></tr>`).join('') || `<tr><td colspan="3" class="muted" style="text-align:center;padding:1.5rem">Sin repuestos</td></tr>`}</tbody>
          </table>
        </div>
      </div>
      ${o.facturaId ? `<div style="margin-top:1rem;padding:1rem 1.2rem;background:var(--success-dim);border:1px solid rgba(76,175,114,0.2);border-radius:10px;display:flex;align-items:center;gap:0.7rem"><i class="bi bi-receipt" style="color:var(--success)"></i><span style="font-size:0.85rem">Factura <strong style="color:var(--success)">#${o.facturaId}</strong> — Total: <strong>${fmt(o.totalFacturado)}</strong> — ${o.estadoPago}</span></div>` : ''}`;
  } catch(err) {
    document.getElementById('detalle-orden-body').innerHTML =
      `<div style="padding:2rem;color:var(--danger);text-align:center">${err.message}</div>`;
  }
}
async function cargarSelectServicios() {
  const s = await apiFetch('/api/servicios');
  document.getElementById('as-servicioid').innerHTML =
    s.map(x => `<option value="${x.servicioId}">${x.nombre} — ${fmt(x.precioBase)}</option>`).join('');
}
async function cargarSelectRepuestos() {
  const r = await apiFetch('/api/repuestos');
  document.getElementById('ar-repuestoid').innerHTML =
    r.map(x => `<option value="${x.repuestoId}">${x.nombre} (Stock: ${x.stockActual}) — ${fmt(x.precioVenta)}</option>`).join('');
}
async function guardarServicioOrden() {
  try {
    await apiFetch(`/api/ordenes/${_ordenActual}/servicios`, { method: 'POST', body: JSON.stringify({
      servicioId: +document.getElementById('as-servicioid').value,
      cantidad:   +document.getElementById('as-cantidad').value,
      manoObra:   +document.getElementById('as-mano').value
    })});
    toast('Servicio agregado'); closeModal('modal-add-servicio'); verOrden(_ordenActual);
  } catch(err) { toast(err.message, 'error'); }
}
async function guardarRepuestoOrden() {
  try {
    await apiFetch(`/api/ordenes/${_ordenActual}/repuestos`, { method: 'POST', body: JSON.stringify({
      repuestoId: +document.getElementById('ar-repuestoid').value,
      cantidad:   +document.getElementById('ar-cantidad').value
    })});
    toast('Repuesto agregado'); closeModal('modal-add-repuesto'); verOrden(_ordenActual);
  } catch(err) { toast(err.message, 'error'); }
}
function abrirCierreOrden() { openModal('modal-cerrar-orden'); }
async function cerrarOrden() {
  try {
    const res = await apiFetch(`/api/ordenes/${_ordenActual}/cerrar`, { method: 'POST', body: JSON.stringify({
      observaciones: document.getElementById('co-obs').value,
      metodoPago:    document.getElementById('co-metodo').value
    })});
    toast(`Orden cerrada — Factura #${res.facturaId} — Total: ${fmt(res.total)}`);
    closeModal('modal-cerrar-orden'); closeModal('modal-detalle-orden'); cargarOrdenes();
  } catch(err) { toast(err.message, 'error'); }
}

// ─── REPUESTOS ───
async function cargarRepuestos() {
  document.getElementById('tabla-repuestos').innerHTML = loading();
  try {
    const data = await apiFetch('/api/repuestos');
    if (!data.length) { document.getElementById('tabla-repuestos').innerHTML = empty(); return; }
    document.getElementById('tabla-repuestos').innerHTML =
      `<table class="data-table">
        <thead><tr><th>Referencia</th><th>Nombre</th><th>Proveedor</th><th>P. Venta</th><th>Stock</th><th>Min</th></tr></thead>
        <tbody>${data.map(r => {
          const low = r.stockActual <= r.stockMinimo;
          return `<tr>
            <td><code style="color:var(--text-2);font-size:0.78rem">${r.referencia}</code></td>
            <td>${r.nombre}</td>
            <td class="muted">${r.nombreProveedor ?? '—'}</td>
            <td>${fmt(r.precioVenta)}</td>
            <td style="${low ? 'color:var(--danger);font-weight:600' : ''}">${r.stockActual}</td>
            <td class="muted">${r.stockMinimo}</td>
          </tr>`;
        }).join('')}</tbody>
      </table>`;
  } catch(err) { toast(err.message, 'error'); }
}
async function guardarRepuesto() {
  try {
    await apiFetch('/api/repuestos', { method: 'POST', body: JSON.stringify({
      proveedorId:  +document.getElementById('r-proveedorid').value,
      referencia:    document.getElementById('r-referencia').value,
      nombre:        document.getElementById('r-nombre').value,
      descripcion:   document.getElementById('r-descripcion').value,
      precioCompra: +document.getElementById('r-pcompra').value,
      precioVenta:  +document.getElementById('r-pventa').value,
      stockActual:  +document.getElementById('r-stock').value,
      stockMinimo:  +document.getElementById('r-stockmin').value
    })});
    toast('Repuesto registrado'); closeModal('modal-repuesto'); cargarRepuestos();
  } catch(err) { toast(err.message, 'error'); }
}

// ─── CDA ───
async function cargarCDA() {
  document.getElementById('tabla-cda').innerHTML = loading();
  try {
    const data = await apiFetch('/api/revisionescda');
    if (!data.length) { document.getElementById('tabla-cda').innerHTML = empty(); return; }
    document.getElementById('tabla-cda').innerHTML =
      `<table class="data-table">
        <thead><tr><th>Placa</th><th>Vehiculo</th><th>Cliente</th><th>Inspector</th><th>Fecha</th><th>Resultado</th><th>Vencimiento</th></tr></thead>
        <tbody>${data.map(r => `<tr>
          <td><code style="color:var(--gold);font-size:0.78rem">${r.placa ?? '—'}</code></td>
          <td>${r.vehiculoDesc ?? '—'}</td>
          <td class="muted">${r.cliente ?? '—'}</td>
          <td class="muted">${r.inspector ?? '—'}</td>
          <td class="muted">${fmtDate(r.fechaRevision)}</td>
          <td>${badgeEstado(r.resultado)}</td>
          <td class="muted">${fmtDate(r.fechaVencimiento)}</td>
        </tr>`).join('')}</tbody>
      </table>`;
  } catch(err) { toast(err.message, 'error'); }
}
async function guardarCDA() {
  try {
    await apiFetch('/api/revisionescda', { method: 'POST', body: JSON.stringify({
      vehiculoId:   +document.getElementById('cda-vehiculoid').value,
      empleadoId:   +document.getElementById('cda-empleadoid').value,
      resultado:     document.getElementById('cda-resultado').value,
      observaciones: document.getElementById('cda-obs').value
    })});
    toast('Revision CDA registrada'); closeModal('modal-cda'); cargarCDA();
  } catch(err) { toast(err.message, 'error'); }
}

// ─── INIT ───
async function inicializar() {
  try {
    const empleados = await apiFetch('/api/empleados');
    const opts = empleados.map(e => `<option value="${e.empleadoId}">${e.nombre} ${e.apellido} — ${e.nombreCargo}</option>`).join('');
    document.getElementById('o-empleadoid').innerHTML = opts;
    document.getElementById('cda-empleadoid').innerHTML = opts;
  } catch(_) {}
  cargarDashboard();
}
inicializar();
