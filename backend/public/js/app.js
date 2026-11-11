const API = 'http://localhost:8080';

function fmt(n) {
  if (n == null) return '—';
  return new Intl.NumberFormat('es-CO', { style: 'currency', currency: 'COP', maximumFractionDigits: 0 }).format(n);
}
function fmtDate(d) {
  if (!d) return '—';

  if (typeof d === 'string') {
    const match = d.match(/^(\d{4})-(\d{2})-(\d{2})/);

    if (match) {
      const [, anio, mes, dia] = match;
      return `${dia}/${mes}/${anio}`;
    }
  }

  return new Date(d).toLocaleDateString('es-CO', {
    timeZone: 'America/Bogota'
  });
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
    ({
      dashboard: cargarDashboard, clientes: cargarClientes, vehiculos: cargarVehiculos,
      ordenes: cargarOrdenes, repuestos: cargarRepuestos, cda: cargarCDA
    })[link.dataset.section]?.();
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
      ordenes.filter(o => ['PENDIENTE', 'EN_PROCESO'].includes(o.estado)).length;
    await cargarEstadisticasHistorial();

    document.getElementById('dashboard-ordenes').innerHTML = ordenes.length === 0
      ? empty('Sin ordenes recientes')
      : `<table class="data-table"><thead><tr><th>#</th><th>Cliente</th><th>Placa</th><th>Estado</th><th>Total</th></tr></thead><tbody>
          ${ordenes.slice(0, 8).map(o => `<tr>
            <td style="color:var(--text-2)">${o.ordenId}</td>
            <td>${o.cliente ?? '—'}</td>
            <td><code style="color:var(--gold);font-size:0.78rem">${o.placa ?? '—'}</code></td>
            <td>${badgeEstado(o.estado)}</td>
            <td>${fmt(o.totalFacturado)}</td>
          </tr>`).join('')}
         </tbody></table>`;

    document.getElementById('dashboard-stock').innerHTML = stock.length === 0
      ? `<div class="empty-state" style="color:var(--success)">Todo el stock OK</div>`
      : `<table class="data-table"><thead><tr><th>Repuesto</th><th>Stock</th><th>Min</th><th>Proveedores</th></tr></thead><tbody>
          ${stock.map(r => `<tr>
            <td>${r.nombre ?? r.Nombre}</td>
            <td style="color:var(--danger);font-weight:600">${r.stock_actual ?? r.StockActual}</td>
            <td class="muted">${r.stock_minimo ?? r.StockMinimo}</td>
          </tr>`).join('')}
         </tbody></table>`;
  } catch (err) { toast(err.message, 'error'); }
}

// ─── ESTADISTICAS HISTORIAL TECNICO ───

async function cargarEstadisticasHistorial() {
  const totalEventos = document.getElementById('stat-historial-eventos');
  const vehiculosAtendidos = document.getElementById('stat-historial-vehiculos');
  const contenedorTipos = document.getElementById('dashboard-historial-tipos');

  try {
    const estadisticas = await apiFetch('/api/historial/estadisticas');

    // Total de eventos técnicos
    totalEventos.textContent = estadisticas.totalEventos ?? 0;

    // Vehículos atendidos
    vehiculosAtendidos.textContent = estadisticas.vehiculosAtendidos ?? 0;

    // Eventos agrupados por tipo
    const porTipo = estadisticas.porTipo || [];

    if (porTipo.length === 0) {
      contenedorTipos.innerHTML = empty('No hay eventos técnicos registrados');
      return;
    }

    const nombresTipos = {
      DIAGNOSTICO: 'Diagnóstico',
      MANTENIMIENTO: 'Mantenimiento',
      REPARACION: 'Reparación',
      INSPECCION_CDA: 'Inspección CDA'
    };

    contenedorTipos.innerHTML = `
      <table class="data-table">
        <thead>
          <tr>
            <th>Tipo de evento</th>
            <th>Cantidad</th>
          </tr>
        </thead>
        <tbody>
          ${porTipo.map(item => `
            <tr>
              <td>${nombresTipos[item.tipoEvento] || item.tipoEvento.replace(/_/g, ' ')}</td>
              <td>
                <strong style="color:var(--gold)">
                  ${item.cantidad}
                </strong>
              </td>
            </tr>
          `).join('')}
        </tbody>
      </table>
    `;

  } catch (err) {
    totalEventos.textContent = '—';
    vehiculosAtendidos.textContent = '—';

    contenedorTipos.innerHTML = empty(
      'No fue posible cargar las estadísticas del historial técnico'
    );

    console.error('Error cargando estadísticas de MongoDB:', err);
  }
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
  } catch (err) { toast(err.message, 'error'); }
}
async function guardarCliente() {
  try {
    await apiFetch('/api/clientes', {
      method: 'POST', body: JSON.stringify({
        cedula: document.getElementById('c-cedula').value,
        nombre: document.getElementById('c-nombre').value,
        apellido: document.getElementById('c-apellido').value,
        telefono: document.getElementById('c-telefono').value,
        email: document.getElementById('c-email').value,
        direccion: document.getElementById('c-direccion').value,
        placa: document.getElementById('c-placa').value,
        marca: document.getElementById('c-marca').value,
        modelo: document.getElementById('c-modelo').value,
        anio: +document.getElementById('c-anio').value,
        color: document.getElementById('c-color').value,
        cilindraje: +document.getElementById('c-cilindraje').value || null,
        kmActual: +document.getElementById('c-km').value
      })
    });
    toast('Cliente registrado'); closeModal('modal-cliente'); cargarClientes();
  } catch (err) { toast(err.message, 'error'); }
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
        <thead><tr><th>Placa</th><th>Marca</th><th>Modelo</th><th>Año</th><th>Km</th><th>Cliente</th><th>Historial</th></tr></thead>
        <tbody>${data.map(v => `<tr>
          <td><code style="color:var(--gold);font-size:0.78rem">${v.placa}</code></td>
          <td>${v.marca}</td>
          <td>${v.modelo}</td>
          <td class="muted">${v.anio}</td>
          <td class="muted">${(v.kmActual || 0).toLocaleString()} km</td>
          <td class="muted">${v.nombreCliente ?? '—'}</td>
          <td><button class="btn btn-ghost btn-sm" onclick="abrirHistorial(${v.vehiculoId}, '${v.placa}', '${v.marca} ${v.modelo}')"><i class="bi bi-clock-history"></i> Historial</button></td>
        </tr>`).join('')}</tbody>
      </table>`;
  } catch (err) { toast(err.message, 'error'); }
}
async function guardarVehiculo() {
  try {
    await apiFetch('/api/vehiculos', {
      method: 'POST', body: JSON.stringify({
        clienteId: +document.getElementById('v-clienteid').value,
        placa: document.getElementById('v-placa').value,
        marca: document.getElementById('v-marca').value,
        modelo: document.getElementById('v-modelo').value,
        anio: +document.getElementById('v-anio').value,
        color: document.getElementById('v-color').value,
        kmActual: +document.getElementById('v-km').value
      })
    });
    toast('Vehiculo registrado'); closeModal('modal-vehiculo'); cargarVehiculos();
  } catch (err) { toast(err.message, 'error'); }
}

// ─── HISTORIAL TECNICO ───

let historialVehiculoActual = null;
let historialEventoEditando = null;
let empleadosDisponibles = [];

async function abrirHistorial(vehiculoId, placa, descripcion) {
  historialVehiculoActual = vehiculoId;
  historialEventoEditando = null;

  document.getElementById('historial-vehiculo-info').textContent =
    `${placa} · ${descripcion} · ID ${vehiculoId}`;

  limpiarFormularioHistorial();
  ocultarFormularioHistorial();

  openModal('modal-historial');
  await cargarHistorialVehiculo();
}

async function cargarHistorialVehiculo() {
  const cont = document.getElementById('lista-historial');
  cont.innerHTML = loading();

  try {
    const data = await apiFetch(
      `/api/historial/vehiculo/${historialVehiculoActual}`
    );

    if (!data || data.length === 0) {
      cont.innerHTML = empty(
        'Este vehículo todavía no tiene eventos técnicos registrados'
      );
      return;
    }

    cont.innerHTML = `
      <div class="table-responsive">
        <table class="data-table">
          <thead>
            <tr>
              <th>Fecha</th>
              <th>Tipo</th>
              <th>Técnico</th>
              <th>Detalle</th>
              <th>Km</th>
              <th>Observaciones</th>
              <th>Acciones</th>
            </tr>
          </thead>
          <tbody>
            ${data.map(item => `
              <tr>
                <td class="muted">${fmtDate(item.fechaEvento)}</td>
                <td>${String(item.tipoEvento).replace(/_/g, ' ')}</td>
                <td>
                  ${item.tecnico?.nombre ?? '—'}<br>
                  <small class="muted">
                    ${item.tecnico?.cargo ?? ''}
                  </small>
                </td>
                <td>${item.detalle?.descripcion ?? '—'}</td>
                <td class="muted">
                  ${item.detalle?.kilometraje ?? '—'}
                </td>
                <td class="muted">
                  ${item.observaciones ?? '—'}
                </td>
                <td>
                  <div class="actions">
                    <button
                      class="btn btn-icon"
                      title="Editar evento"
                      onclick="editarHistorial('${item.id}')">
                      <i class="bi bi-pencil"></i>
                    </button>

                    <button
                      class="btn btn-icon"
                      title="Eliminar evento"
                      style="color:var(--danger)"
                      onclick="eliminarHistorial('${item.id}')">
                      <i class="bi bi-trash"></i>
                    </button>
                  </div>
                </td>
              </tr>
            `).join('')}
          </tbody>
        </table>
      </div>
    `;

  } catch (err) {
    cont.innerHTML = empty(err.message);
    toast(err.message, 'error');
  }
}

function actualizarCargoTecnico() {
  const empleadoId = Number(document.getElementById('h-empleado').value);
  const empleado = empleadosDisponibles.find(e => Number(e.empleadoId) === empleadoId);
  document.getElementById('h-cargo').value = empleado?.nombreCargo ?? '';
}

async function cargarEmpleadosHistorial() {
  const select = document.getElementById('h-empleado');
  select.innerHTML = '<option value="">Cargando empleados...</option>';

  try {
    empleadosDisponibles = await apiFetch('/api/empleados');
    select.innerHTML = '<option value="">Selecciona un empleado</option>' +
      empleadosDisponibles.map(e => `
        <option value="${e.empleadoId}">
          ${e.nombre} ${e.apellido} — ${e.nombreCargo}
        </option>
      `).join('');
    actualizarCargoTecnico();
  } catch (err) {
    select.innerHTML = '<option value="">No fue posible cargar empleados</option>';
    toast('No fue posible cargar los empleados: ' + err.message, 'error');
  }
}

function mostrarFormularioHistorial() {
  historialEventoEditando = null;
  limpiarFormularioHistorial();

  document.querySelector('#historial-formulario h6').textContent =
    'REGISTRAR EVENTO TECNICO';

  document.querySelector(
    '#historial-formulario button.btn-primary'
  ).textContent = 'Guardar evento';

  document.getElementById('historial-formulario').classList.remove('d-none');
}

function ocultarFormularioHistorial() {
  document.getElementById('historial-formulario').classList.add('d-none');
}

function limpiarFormularioHistorial() {
  document.getElementById('h-tipo').value = 'DIAGNOSTICO';
  document.getElementById('h-fecha').value =
    new Date().toISOString().slice(0, 10);
  document.getElementById('h-empleado').value = '';
  document.getElementById('h-cargo').value = '';
  document.getElementById('h-km').value = '';
  document.getElementById('h-detalle').value = '';
  document.getElementById('h-observaciones').value = '';
  document.getElementById('h-recomendaciones').value = '';
}

async function editarHistorial(id) {
  try {
    const evento = await apiFetch(`/api/historial/${id}`);

    historialEventoEditando = id;

    document.getElementById('h-tipo').value =
      evento.tipoEvento ?? 'DIAGNOSTICO';

    document.getElementById('h-fecha').value =
      evento.fechaEvento
        ? new Date(evento.fechaEvento).toISOString().slice(0, 10)
        : '';

    document.getElementById('h-empleado').value =
      evento.tecnico?.empleadoId ?? '';

    actualizarCargoTecnico();

    // Los registros antiguos sin empleadoId requieren seleccionar un empleado vigente.
    if (!evento.tecnico?.empleadoId) {
      toast('Selecciona el empleado responsable para actualizar este registro', 'error');
    }

    document.getElementById('h-km').value =
      evento.detalle?.kilometraje ?? '';

    document.getElementById('h-detalle').value =
      evento.detalle?.descripcion ?? '';

    document.getElementById('h-observaciones').value =
      evento.observaciones ?? '';

    document.getElementById('h-recomendaciones').value =
      evento.recomendaciones ?? '';

    document.querySelector('#historial-formulario h6').textContent =
      'EDITAR EVENTO TECNICO';

    document.querySelector(
      '#historial-formulario button.btn-primary'
    ).textContent = 'Actualizar evento';

    document.getElementById('historial-formulario').classList.remove('d-none');

  } catch (err) {
    toast(err.message, 'error');
  }
}

async function guardarHistorial() {
  try {
    const fecha = document.getElementById('h-fecha').value;
    const descripcion =
      document.getElementById('h-detalle').value.trim();
    const empleadoId = Number(document.getElementById('h-empleado').value);

    if (!fecha || !descripcion || !empleadoId) {
      toast('Selecciona el técnico y completa la fecha y el detalle del evento', 'error');
      return;
    }

    const datos = {
      vehiculoId: historialVehiculoActual,
      tipoEvento: document.getElementById('h-tipo').value,
      fechaEvento: `${fecha}T12:00:00Z`,
      tecnico: {
        empleadoId
      },
      detalle: {
        descripcion,
        kilometraje:
          Number(document.getElementById('h-km').value) || null
      },
      observaciones:
        document.getElementById('h-observaciones').value.trim(),
      recomendaciones:
        document.getElementById('h-recomendaciones').value.trim()
    };

    if (historialEventoEditando) {
      await apiFetch(`/api/historial/${historialEventoEditando}`, {
        method: 'PUT',
        body: JSON.stringify(datos)
      });

      toast('Evento técnico actualizado correctamente');

    } else {
      await apiFetch('/api/historial', {
        method: 'POST',
        body: JSON.stringify(datos)
      });

      toast('Evento técnico registrado en MongoDB');
    }

    historialEventoEditando = null;
    limpiarFormularioHistorial();
    ocultarFormularioHistorial();

    await cargarHistorialVehiculo();

  } catch (err) {
    toast(err.message, 'error');
  }
}

async function eliminarHistorial(id) {
  const confirmar = confirm(
    '¿Estás seguro de eliminar este evento del historial técnico? Esta acción no se puede deshacer.'
  );

  if (!confirmar) return;

  try {
    await apiFetch(`/api/historial/${id}`, {
      method: 'DELETE'
    });

    if (historialEventoEditando === id) {
      historialEventoEditando = null;
      limpiarFormularioHistorial();
      ocultarFormularioHistorial();
    }

    toast('Evento técnico eliminado correctamente');

    await cargarHistorialVehiculo();

  } catch (err) {
    toast(err.message, 'error');
  }
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
  } catch (err) { toast(err.message, 'error'); }
}
async function guardarOrden() {
  try {
    await apiFetch('/api/ordenes', {
      method: 'POST', body: JSON.stringify({
        vehiculoId: +document.getElementById('o-vehiculoid').value,
        empleadoId: +document.getElementById('o-empleadoid').value,
        kmIngreso: +document.getElementById('o-km').value,
        fechaEstimada: document.getElementById('o-fecha').value || null,
        diagnostico: document.getElementById('o-diagnostico').value
      })
    });
    toast('Orden creada'); closeModal('modal-orden'); cargarOrdenes();
  } catch (err) { toast(err.message, 'error'); }
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
    const editable = ['PENDIENTE', 'EN_PROCESO'].includes(o.estado);
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
  } catch (err) {
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
    await apiFetch(`/api/ordenes/${_ordenActual}/servicios`, {
      method: 'POST', body: JSON.stringify({
        servicioId: +document.getElementById('as-servicioid').value,
        cantidad: +document.getElementById('as-cantidad').value,
        manoObra: +document.getElementById('as-mano').value
      })
    });
    toast('Servicio agregado'); closeModal('modal-add-servicio'); verOrden(_ordenActual);
  } catch (err) { toast(err.message, 'error'); }
}
async function guardarRepuestoOrden() {
  try {
    await apiFetch(`/api/ordenes/${_ordenActual}/repuestos`, {
      method: 'POST', body: JSON.stringify({
        repuestoId: +document.getElementById('ar-repuestoid').value,
        cantidad: +document.getElementById('ar-cantidad').value
      })
    });
    toast('Repuesto agregado'); closeModal('modal-add-repuesto'); verOrden(_ordenActual);
  } catch (err) { toast(err.message, 'error'); }
}
function abrirCierreOrden() { openModal('modal-cerrar-orden'); }
async function cerrarOrden() {
  try {
    const res = await apiFetch(`/api/ordenes/${_ordenActual}/cerrar`, {
      method: 'POST', body: JSON.stringify({
        observaciones: document.getElementById('co-obs').value,
        metodoPago: document.getElementById('co-metodo').value
      })
    });
    toast(`Orden cerrada — Factura #${res.facturaId} — Total: ${fmt(res.total)}`);
    closeModal('modal-cerrar-orden'); closeModal('modal-detalle-orden'); cargarOrdenes();
  } catch (err) { toast(err.message, 'error'); }
}

// ─── REPUESTOS ───
async function cargarRepuestos() {
  document.getElementById('tabla-repuestos').innerHTML = loading();
  try {
    const data = await apiFetch('/api/repuestos');
    if (!data.length) { document.getElementById('tabla-repuestos').innerHTML = empty(); return; }
    document.getElementById('tabla-repuestos').innerHTML =
      `<table class="data-table">
        <thead><tr><th>Referencia</th><th>Nombre</th><th>Proveedor</th><th>P. Venta</th><th>Stock</th><th>Min</th><th>Proveedores</th></tr></thead>
        <tbody>${data.map(r => {
        const low = r.stockActual <= r.stockMinimo;
        return `<tr>
            <td><code style="color:var(--text-2);font-size:0.78rem">${r.referencia}</code></td>
            <td>${r.nombre}</td>
            <td class="muted">${r.nombreProveedor ?? '—'}</td>
            <td>${fmt(r.precioVenta)}</td>
            <td style="${low ? 'color:var(--danger);font-weight:600' : ''}">${r.stockActual}</td>
            <td class="muted">${r.stockMinimo}</td>
            <td><button class="btn btn-ghost btn-sm" onclick="abrirProveedoresRepuesto(${r.repuestoId})">Gestionar</button></td>
          </tr>`;
      }).join('')}</tbody>
      </table>`;
  } catch (err) { toast(err.message, 'error'); }
}
async function guardarRepuesto() {
  try {
    await apiFetch('/api/repuestos', {
      method: 'POST', body: JSON.stringify({
        proveedorId: +document.getElementById('r-proveedorid').value,
        referencia: document.getElementById('r-referencia').value,
        nombre: document.getElementById('r-nombre').value,
        descripcion: document.getElementById('r-descripcion').value,
        precioCompra: +document.getElementById('r-pcompra').value,
        precioVenta: +document.getElementById('r-pventa').value,
        stockActual: +document.getElementById('r-stock').value,
        stockMinimo: +document.getElementById('r-stockmin').value
      })
    });
    toast('Repuesto registrado'); closeModal('modal-repuesto'); cargarRepuestos();
  } catch (err) { toast(err.message, 'error'); }
}

// ─── PROVEEDORES POR REPUESTO ───
async function abrirProveedoresRepuesto(repuestoId) {
  document.getElementById('rp-repuestoid').value = repuestoId;
  try {
    const proveedores = await apiFetch('/api/proveedores');
    document.getElementById('rp-proveedor').innerHTML = proveedores.map(p => `<option value="${p.proveedorId}">${p.nombre}</option>`).join('');
    await cargarProveedoresRepuesto();
    openModal('modal-proveedores-repuesto');
  } catch (err) { toast(err.message, 'error'); }
}
async function cargarProveedoresRepuesto() {
  const id = document.getElementById('rp-repuestoid').value;
  const data = await apiFetch(`/api/repuestos/${id}/proveedores`);
  document.getElementById('rp-lista').innerHTML = data.length ? `
    <table class="data-table"><thead><tr><th>Proveedor</th><th>Precio compra</th><th>Entrega</th><th></th></tr></thead>
    <tbody>${data.map(p => `<tr><td>${p.proveedor}</td><td>${fmt(p.precioCompra)}</td><td>${p.diasEntrega} días</td>
    <td><button class="btn btn-ghost btn-sm" onclick="eliminarProveedorRepuesto(${p.proveedorId})">Quitar</button></td></tr>`).join('')}</tbody></table>` : empty('Sin proveedores asociados');
}
async function guardarProveedorRepuesto() {
  try {
    const id = +document.getElementById('rp-repuestoid').value;
    await apiFetch(`/api/repuestos/${id}/proveedores`, {method:'POST', body:JSON.stringify({
      proveedorId:+document.getElementById('rp-proveedor').value,
      precioCompra:+document.getElementById('rp-precio').value,
      diasEntrega:+document.getElementById('rp-dias').value
    })});
    // Limpiar el formulario después de guardar para evitar repetir datos por accidente.
    document.getElementById('rp-precio').value = '';
    document.getElementById('rp-dias').value = '2';
    const proveedorSelect = document.getElementById('rp-proveedor');
    if (proveedorSelect) proveedorSelect.selectedIndex = 0;
    toast('Proveedor asociado'); await cargarProveedoresRepuesto(); cargarRepuestos();
  } catch(err) { toast(err.message,'error'); }
}
async function eliminarProveedorRepuesto(proveedorId) {
  try {
    const id=+document.getElementById('rp-repuestoid').value;
    await apiFetch(`/api/repuestos/${id}/proveedores?proveedorId=${proveedorId}`,{method:'DELETE'});
    toast('Proveedor eliminado'); await cargarProveedoresRepuesto(); cargarRepuestos();
  } catch(err) { toast(err.message,'error'); }
}

// ─── CDA ───
let revisionCDAEnEdicion = null;

// Las observaciones se registran únicamente por elemento del checklist.
// Ocultamos y retiramos el campo general heredado del HTML.
function ocultarObservacionesGeneralesCDA() {
  const campo = document.getElementById('cda-obs');
  if (!campo) return;
  const grupo = campo.closest('.form-group');
  if (grupo) {
    grupo.remove();
    return;
  }
  const etiqueta = document.querySelector('label[for="cda-obs"]');
  etiqueta?.remove();
  campo.remove();
}

async function cargarCDA() {
  document.getElementById('tabla-cda').innerHTML = loading();
  try {
    const data = await apiFetch('/api/revisionescda');
    if (!data.length) { document.getElementById('tabla-cda').innerHTML = empty(); return; }
    document.getElementById('tabla-cda').innerHTML =
      `<table class="data-table">
        <thead><tr><th>Placa</th><th>Vehiculo</th><th>Cliente</th><th>Inspector</th><th>Fecha</th><th>Resultado</th><th>Vencimiento</th><th>Acciones</th></tr></thead>
        <tbody>${data.map(r => `<tr>
          <td><code style="color:var(--gold);font-size:0.78rem">${r.placa ?? '—'}</code></td>
          <td>${r.vehiculoDesc ?? '—'}</td>
          <td class="muted">${r.cliente ?? '—'}</td>
          <td class="muted">${r.inspector ?? '—'}</td>
          <td class="muted">${fmtDate(r.fechaRevision)}</td>
          <td>${badgeEstado(r.resultado)}</td>
          <td class="muted">${fmtDate(r.fechaVencimiento)}</td>
          <td class="cda-actions"><button class="btn btn-ghost btn-sm" onclick="abrirDetalleCDA(${r.revisionId})">Detalle</button> <button class="btn btn-ghost btn-sm" onclick="editarRevisionCDA(${r.revisionId})">Editar</button> <button class="btn btn-danger btn-sm" onclick="eliminarRevisionCDA(${r.revisionId})">Eliminar</button></td>
        </tr>`).join('')}</tbody>
      </table>`;
  } catch (err) { toast(err.message, 'error'); }
}
async function abrirDetalleCDA(revisionId) {
  const cont = document.getElementById('detalle-cda-body');
  document.getElementById('detalle-cda-titulo').textContent = `REVISION CDA #${revisionId}`;
  cont.innerHTML = loading();
  openModal('modal-detalle-cda');
  try {
    const detalles = await apiFetch(`/api/revisionescda/${revisionId}/detalle`);
    if (!detalles.length) { cont.innerHTML = empty('Esta revisión no tiene checklist registrado'); return; }
    cont.innerHTML = `<table class="data-table"><thead><tr><th>Grupo</th><th>Elemento inspeccionado</th><th>Critico</th><th>Resultado</th><th>Observacion</th></tr></thead><tbody>
      ${detalles.map(d => `<tr><td>${d.grupo ?? '—'}</td><td>${d.nombre ?? '—'}</td><td>${d.esCritico ? 'Sí' : 'No'}</td><td>${d.resultado === 'CONFORME' ? '<span class="badge badge-aprobado">Conforme</span>' : '<span class="badge badge-reprobado">No conforme</span>'}</td><td>${d.observacion || '—'}</td></tr>`).join('')}
    </tbody></table>`;
  } catch (err) { cont.innerHTML = `<div class="text-danger">${err.message}</div>`; }
}

async function editarRevisionCDA(revisionId) {
  try {
    const data = await apiFetch('/api/revisionescda');
    const revision = data.find(r => Number(r.revisionId) === Number(revisionId));
    if (!revision) throw new Error('No se encontró la revisión CDA.');
    revisionCDAEnEdicion = Number(revisionId);
    document.querySelector('#modal-cda .modal-title').textContent = `EDITAR REVISION CDA #${revisionId}`;
    document.querySelector('#modal-cda .modal-footer .btn-primary').textContent = 'Guardar cambios';
    document.getElementById('cda-vehiculoid').value = revision.vehiculoId;
    document.getElementById('cda-empleadoid').value = revision.empleadoId;
    document.getElementById('cda-resultado').value = revision.resultado;
    await cargarChecklistCDA();
    const detalles = await apiFetch(`/api/revisionescda/${revisionId}/detalle`);
    detalles.forEach(d => {
      const resultado = document.querySelector(`.cda-item-resultado[data-item-id="${d.itemId}"]`);
      const observacion = document.querySelector(`.cda-item-obs[data-item-id="${d.itemId}"]`);
      if (resultado) resultado.value = d.resultado;
      if (observacion) observacion.value = d.observacion || '';
    });
    openModal('modal-cda');
  } catch (err) { toast(err.message, 'error'); }
}

async function eliminarRevisionCDA(revisionId) {
  if (!confirm(`¿Seguro que deseas eliminar la revisión CDA #${revisionId} y todo su checklist? Esta acción no se puede deshacer.`)) return;
  try {
    await apiFetch(`/api/revisionescda/${revisionId}`, { method: 'DELETE' });
    toast('Revisión CDA eliminada correctamente');
    await cargarCDA();
  } catch (err) { toast(err.message, 'error'); }
}

async function cargarChecklistCDA() {
  const cont = document.getElementById('cda-checklist');
  try {
    const items = await apiFetch('/api/items-inspeccion');
    cont.innerHTML = items.map(item => `
      <div class="form-row" style="--cols:3;align-items:center;border-bottom:1px solid var(--border);padding:8px 0">
        <div class="form-group"><strong>${item.nombre}</strong><br><small class="muted">${item.grupo}${item.esCritico ? ' · Crítico' : ''}</small></div>
        <div class="form-group"><select class="form-select cda-item-resultado" data-item-id="${item.itemId}">
          <option value="CONFORME">Conforme</option><option value="NO_CONFORME">No conforme</option>
        </select></div>
        <div class="form-group"><input class="form-control cda-item-obs" data-item-id="${item.itemId}" placeholder="Observación"></div>
      </div>`).join('');
  } catch (err) { cont.innerHTML = `<small class="text-danger">${err.message}</small>`; }
}
function limpiarFormularioCDA() {
  const vehiculo = document.getElementById('cda-vehiculoid');
  const inspector = document.getElementById('cda-empleadoid');
  const resultado = document.getElementById('cda-resultado');
  const checklist = document.getElementById('cda-checklist');

  // Cada revisión nueva comienza sin datos de la revisión anterior.
  if (vehiculo) vehiculo.value = '';
  if (inspector) inspector.value = '';
  if (resultado) {
    resultado.value = [...resultado.options].some(o => o.value === 'APROBADO')
      ? 'APROBADO'
      : (resultado.options[0]?.value || '');
  }
  if (checklist) checklist.innerHTML = '';

  revisionCDAEnEdicion = null;
  const titulo = document.querySelector('#modal-cda .modal-title');
  const boton = document.querySelector('#modal-cda .modal-footer .btn-primary');
  if (titulo) titulo.textContent = 'NUEVA REVISION CDA';
  if (boton) boton.textContent = 'Guardar';
}

document.getElementById('modal-cda')?.addEventListener('show.bs.modal', async () => {
  ocultarObservacionesGeneralesCDA();
  if (!revisionCDAEnEdicion) {
    limpiarFormularioCDA();
    await cargarChecklistCDA();
  }
});

async function guardarCDA() {
  try {
    ocultarObservacionesGeneralesCDA();

    const detalles = [...document.querySelectorAll('.cda-item-resultado')].map(sel => ({
      itemId: Number(sel.dataset.itemId),
      resultado: sel.value,
      observacion: document.querySelector(
        `.cda-item-obs[data-item-id="${sel.dataset.itemId}"]`
      )?.value.trim() || ''
    }));

    if (detalles.length === 0) {
      toast('No se encontraron elementos del checklist para guardar', 'error');
      return;
    }

    const payload = {
      vehiculoId: Number(document.getElementById('cda-vehiculoid').value),
      empleadoId: Number(document.getElementById('cda-empleadoid').value),
      resultado: document.getElementById('cda-resultado').value
    };

    if (!payload.vehiculoId || !payload.empleadoId) {
      toast('Selecciona el vehículo y el inspector', 'error');
      return;
    }

    if (revisionCDAEnEdicion) {
      // La revisión y todos sus elementos se actualizan en una sola transacción.
      await apiFetch(`/api/revisionescda/${revisionCDAEnEdicion}`, {
        method: 'PUT',
        body: JSON.stringify({ ...payload, detalles })
      });
      toast('Revisión CDA y checklist actualizados correctamente');
    } else {
      // En creación, primero se genera la revisión y luego se registran sus elementos.
      const revision = await apiFetch('/api/revisionescda', {
        method: 'POST',
        body: JSON.stringify(payload)
      });
      const revisionId = revision.revisionId;
      if (!revisionId) throw new Error('El servidor no devolvió el ID de la revisión creada');

      await apiFetch(`/api/revisionescda/${revisionId}/detalle`, {
        method: 'POST',
        body: JSON.stringify(detalles)
      });
      toast('Revisión CDA y checklist registrados correctamente');
    }

    revisionCDAEnEdicion = null;
    document.querySelector('#modal-cda .modal-title').textContent = 'NUEVA REVISION CDA';
    document.querySelector('#modal-cda .modal-footer .btn-primary').textContent = 'Guardar';
    closeModal('modal-cda');
    await cargarCDA();
  } catch (err) {
    toast(err.message, 'error');
    console.error('Error guardando revisión CDA:', err);
  }
}

document.getElementById('modal-cda')?.addEventListener('hidden.bs.modal', () => {
  limpiarFormularioCDA();
});

// ─── INIT ───
document.addEventListener('change', event => {
  if (event.target?.id === 'h-empleado') actualizarCargoTecnico();
});

async function inicializar() {
  ocultarObservacionesGeneralesCDA();
  try {
    const empleados = await apiFetch('/api/empleados');
    empleadosDisponibles = empleados;
    const opts = empleados.map(e => `<option value="${e.empleadoId}">${e.nombre} ${e.apellido} — ${e.nombreCargo}</option>`).join('');
    document.getElementById('o-empleadoid').innerHTML = opts;
    document.getElementById('cda-empleadoid').innerHTML =
      '<option value="">Selecciona un inspector</option>' + opts;
    const historialSelect = document.getElementById('h-empleado');
    if (historialSelect) {
      historialSelect.innerHTML = '<option value="">Selecciona un empleado</option>' +
        empleados.map(e => `<option value="${e.empleadoId}">${e.nombre} ${e.apellido} — ${e.nombreCargo}</option>`).join('');
    }
  } catch (_) { }
  cargarDashboard();
}
inicializar();
