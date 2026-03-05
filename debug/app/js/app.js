(function () {
  'use strict';

  // ===========================
  // Paths module
  // ===========================
  var Paths = {
    getApiBasePath: function () {
      var path = window.location.pathname;
      if (path.endsWith('/index.html')) path = path.slice(0, -11);
      if (path.endsWith('/')) path = path.slice(0, -1);
      if (path.endsWith('/ui')) return path.slice(0, -3);
      return path;
    },
    buildApiUrl: function (endpoint) {
      if (endpoint[0] !== '/') endpoint = '/' + endpoint;
      return this.getApiBasePath() + endpoint;
    },
    buildWebSocketUrl: function (endpoint) {
      var proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
      if (endpoint[0] !== '/') endpoint = '/' + endpoint;
      return proto + '//' + location.host + this.getApiBasePath() + endpoint;
    }
  };

  // ===========================
  // Formatters module
  // ===========================
  var Fmt = {
    bytes: function (b) {
      if (!b) return '0B';
      var u = ['B', 'KB', 'MB', 'GB', 'TB'], i = 0, s = b;
      while (s >= 1024 && i < u.length - 1) { s /= 1024; i++; }
      return s.toFixed(1) + u[i];
    },
    duration: function (ms) {
      if (!ms) return '0ms';
      if (ms < 1000) return ms + 'ms';
      if (ms < 60000) return (ms / 1000).toFixed(1) + 's';
      if (ms < 3600000) return (ms / 60000).toFixed(1) + 'min';
      return (ms / 3600000).toFixed(1) + 'h';
    },
    time: function (ts) {
      if (!ts) return '';
      var t = ts < 1e10 ? ts * 1000 : ts;
      return new Date(t).toLocaleTimeString();
    },
    dateTime: function (ts) {
      if (!ts) return '';
      var t = ts < 1e10 ? ts * 1000 : ts;
      var d = new Date(t);
      var pad = function (n) { return String(n).padStart(2, '0'); };
      return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()) +
        ' ' + pad(d.getHours()) + ':' + pad(d.getMinutes()) + ':' + pad(d.getSeconds());
    },
    number: function (n) { return n ? n.toLocaleString() : '0'; },
    latency: function (v) {
      if (!v) return 'N/A';
      if (typeof v === 'string') {
        var m = v.match(/^([\d.]+)(ms|s|μs|us|ns)?/);
        if (m) { var unit = m[2] === 'us' ? 'μs' : (m[2] || 'ms'); return parseFloat(m[1]).toFixed(2) + unit; }
        return v;
      }
      return Number(v).toFixed(2) + 'ms';
    },
    method: function (m) { return m ? m.toUpperCase() : 'UNKNOWN'; },
    statusBadge: function (status) {
      var s = String(status);
      var map = {
        active: 'success', running: 'success', completed: 'info', failed: 'danger', blocked: 'warning', waiting: 'secondary',
        Active: 'success', Running: 'success', Completed: 'info', Failed: 'danger', Blocked: 'warning', Waiting: 'secondary',
        '200': 'success', '201': 'success', '204': 'success',
        '301': 'warning', '302': 'warning',
        '400': 'warning', '401': 'warning', '403': 'warning', '404': 'warning',
        '500': 'danger', '502': 'danger', '503': 'danger'
      };
      return 'badge-status-' + (map[s] || map[s.toLowerCase()] || 'secondary');
    },
    methodBadge: function (method) {
      var map = { GET: 'primary', POST: 'success', PUT: 'warning', DELETE: 'danger', PATCH: 'info' };
      return map[(method || '').toUpperCase()] || 'secondary';
    }
  };

  // ===========================
  // API module
  // ===========================
  var Api = {
    _get: function (url, params) {
      var qs = '';
      if (params) {
        var parts = [];
        for (var k in params) { if (params[k] !== undefined && params[k] !== null && params[k] !== '') parts.push(encodeURIComponent(k) + '=' + encodeURIComponent(params[k])); }
        if (parts.length) qs = '?' + parts.join('&');
      }
      return fetch(Paths.buildApiUrl(url) + qs, { headers: { 'Content-Type': 'application/json' } })
        .then(function (r) { if (!r.ok) throw new Error('HTTP ' + r.status); return r.json(); });
    },
    _post: function (url, body) {
      return fetch(Paths.buildApiUrl(url), {
        method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body)
      }).then(function (r) { if (!r.ok) throw new Error('HTTP ' + r.status); return r.json(); });
    },
    getSystemInfo: function () { return this._get('/system'); },
    getMonitorStats: function () { return this._get('/stats'); },
    getGoroutines: function (params) { return this._get('/goroutines', params); },
    getGoroutineDetail: function (id) { return this._get('/goroutine/' + encodeURIComponent(id)); },
    getRequests: function (params) { return this._get('/requests', params); },
    getRequestDetail: function (id) { return this._get('/request/' + encodeURIComponent(id)); },
    getHeapProfile: function () { return this._get('/heap'); }
  };

  // ===========================
  // WebSocket module
  // ===========================
  var WS = {
    ws: null, connected: false, reconnectTimer: null, reconnectDelay: 2000,
    systemStats: {}, recentGoroutines: [], recentRequests: [],
    listeners: [],

    connect: function () {
      var self = this;
      if (self.ws) self.disconnect();
      try {
        self.ws = new WebSocket(Paths.buildWebSocketUrl('/ws'));
        self.ws.onopen = function () {
          self.connected = true;
          self.reconnectDelay = 2000;
          self._updateStatus();
          self.ws.send(JSON.stringify({ type: 'subscribe', data: { subscribe_stats: true, subscribe_goroutines: true, subscribe_requests: true } }));
          self.ws.send(JSON.stringify({ type: 'get_stats' }));
        };
        self.ws.onmessage = function (e) {
          try { self._processMessage(JSON.parse(e.data)); } catch (err) { console.error('WS parse error:', err); }
        };
        self.ws.onerror = function () { self.connected = false; self._updateStatus(); };
        self.ws.onclose = function () {
          self.connected = false;
          self._updateStatus();
          self.reconnectTimer = setTimeout(function () { self.connect(); }, self.reconnectDelay);
          self.reconnectDelay = Math.min(self.reconnectDelay * 1.5, 30000);
        };
      } catch (err) { console.error('WS connect error:', err); }
    },
    disconnect: function () {
      if (this.reconnectTimer) { clearTimeout(this.reconnectTimer); this.reconnectTimer = null; }
      if (this.ws) { this.ws.onclose = null; this.ws.close(); this.ws = null; }
      this.connected = false;
      this._updateStatus();
    },
    _processMessage: function (msg) {
      if (!msg || typeof msg.type !== 'string') return;
      switch (msg.type) {
        case 'stats': case 'stats_update':
          if (msg.data && typeof msg.data === 'object' && !Array.isArray(msg.data)) {
            Object.assign(this.systemStats, msg.data);
          }
          break;
        case 'goroutine': case 'goroutine_update':
          this.recentGoroutines.unshift(msg.data);
          if (this.recentGoroutines.length > 10) this.recentGoroutines.length = 10;
          break;
        case 'request': case 'request_update':
          var r = msg.data;
          r.resp_status_code = parseInt(String(r.resp_status_code)) || 0;
          this.recentRequests.unshift(r);
          if (this.recentRequests.length > 10) this.recentRequests.length = 10;
          break;
      }
      this._notify(msg.type);
    },
    _updateStatus: function () {
      var el = document.getElementById('ws-status');
      if (!el) return;
      if (this.connected) {
        el.className = 'badge rounded-pill bg-success';
        el.innerHTML = '<i class="bi bi-wifi"></i> 已连接';
      } else {
        el.className = 'badge rounded-pill bg-secondary';
        el.innerHTML = '<i class="bi bi-wifi-off"></i> 未连接';
      }
    },
    onChange: function (fn) { this.listeners.push(fn); },
    _notify: function (type) { this.listeners.forEach(function (fn) { fn(type); }); }
  };

  // ===========================
  // HTML helpers
  // ===========================
  function esc(s) {
    if (s === null || s === undefined) return '';
    var d = document.createElement('div');
    d.appendChild(document.createTextNode(String(s)));
    return d.innerHTML.replace(/"/g, '&quot;').replace(/'/g, '&#39;');
  }

  function jsonHighlight(obj) {
    var json = JSON.stringify(obj, null, 2);
    if (!json) return '';
    return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
      var cls = 'json-number';
      if (/^"/.test(match)) { cls = /:$/.test(match) ? 'json-key' : 'json-string'; }
      else if (/true|false/.test(match)) cls = 'json-boolean';
      else if (/null/.test(match)) cls = 'json-null';
      return '<span class="' + cls + '">' + esc(match) + '</span>';
    });
  }

  function tryParseJson(str) {
    if (!str) return null;
    try { return JSON.parse(str); } catch (e) { return null; }
  }

  function renderJsonOrPre(data, rawStr, extraClass) {
    if (typeof data === 'object' && data !== null) {
      return '<div class="json-display">' + jsonHighlight(data) + '</div>';
    }
    return '<pre class="code-block ' + (extraClass || '') + '">' + esc(rawStr || data) + '</pre>';
  }

  // ===========================
  // SessionLogs renderer
  // ===========================
  var SessionLogs = {
    render: function (sessionLogs, showTitle) {
      if (!sessionLogs) return '';
      var logs;
      try {
        logs = JSON.parse(sessionLogs);
        if (!Array.isArray(logs) || logs.length === 0) return typeof sessionLogs === 'string' && sessionLogs.trim() ? '<pre class="code-block">' + esc(sessionLogs) + '</pre>' : '';
      } catch (e) {
        return sessionLogs.trim() ? '<pre class="code-block">' + esc(sessionLogs) + '</pre>' : '';
      }
      var h = '';
      if (showTitle !== false) h += '<h6 class="mb-3 fw-semibold">会话日志</h6>';
      h += '<div class="session-logs-container">';
      logs.forEach(function (log, idx) {
        var logTime = new Date(log.time * 1000).toLocaleString();
        var level = 'INFO';
        if (typeof log.level === 'number') {
          var lvlMap = { '-1': 'DEBUG', 0: 'INFO', 1: 'WARN', 2: 'ERROR', 3: 'ERROR', 4: 'ERROR', 5: 'ERROR' };
          level = lvlMap[log.level] || 'INFO';
        } else {
          level = (log.level || 'INFO').toString().toUpperCase();
        }
        h += '<div class="log-entry">' +
          '<div class="log-header">' +
          '<span class="log-index">#' + (idx + 1) + '</span>' +
          '<span>' + esc(logTime) + '</span>' +
          '<span class="log-level level-' + level.toLowerCase() + '">' + level + '</span>' +
          '<span class="log-caller">' + esc(log.caller || 'Unknown') + '</span>' +
          '</div>' +
          '<div class="ps-2"><code class="log-code">' + esc(log.message || '') + '</code></div>' +
          '</div>';
      });
      h += '</div>';
      return h;
    }
  };

  // ===========================
  // GoroutineModal
  // ===========================
  var GoroutineModal = {
    show: function (goroutine) {
      var g = goroutine;
      var now = Math.floor(Date.now() / 1000);
      var dur = g.end_time ? (g.end_time - g.start_time) : (now - g.start_time);
      var stack = g.stack || g.stack_trace || '';
      var name = this._formatName(g);
      var sessionStr = g.session_logs ? (typeof g.session_logs === 'string' ? g.session_logs : JSON.stringify(g.session_logs)) : null;

      var h = '<div class="modal-header"><h5 class="modal-title">协程详细信息</h5><button type="button" class="btn-close" data-bs-dismiss="modal"></button></div>';
      h += '<div class="modal-body">';
      h += '<table class="table table-sm table-bordered mb-3"><tbody>';
      h += '<tr><th style="width:120px">协程ID</th><td>' + esc(g.id) + '</td><th style="width:120px">状态</th><td><span class="badge ' + Fmt.statusBadge(g.status) + '">' + esc(g.status) + '</span></td></tr>';
      h += '<tr><th>名称</th><td class="fw-medium">' + esc(name) + '</td><th>运行时长</th><td>' + Fmt.duration(dur * 1000) + '</td></tr>';
      h += '<tr><th>开始时间</th><td>' + Fmt.dateTime(g.start_time) + '</td>';
      if (g.end_time) h += '<th>结束时间</th><td>' + Fmt.dateTime(g.end_time) + '</td>';
      else h += '<td colspan="2"></td>';
      h += '</tr>';
      if (g.request_id) h += '<tr><th>关联请求</th><td colspan="3" class="font-mono small">' + esc(g.request_id) + '</td></tr>';
      h += '</tbody></table>';

      if (g.cpu_usage !== undefined || g.memory_usage !== undefined) {
        h += '<div class="card mb-2"><div class="card-header py-1"><strong>性能指标</strong></div><div class="card-body py-2">';
        if (g.cpu_usage !== undefined) h += '<span class="me-3">CPU: ' + g.cpu_usage.toFixed(2) + '%</span>';
        if (g.memory_usage !== undefined) h += '<span>内存: ' + (g.memory_usage / 1024 / 1024).toFixed(2) + ' MB</span>';
        h += '</div></div>';
      }

      if (sessionStr && sessionStr !== '[]') {
        h += '<div class="card mb-2"><div class="card-body py-2">' + SessionLogs.render(sessionStr) + '</div></div>';
      }

      if (stack) {
        h += '<div class="card mb-2"><div class="card-header py-1"><strong>堆栈跟踪</strong></div><div class="card-body py-2"><pre class="code-block mb-0">' + esc(stack) + '</pre></div></div>';
      }

      if (g.error) {
        h += '<div class="card mb-2"><div class="card-header py-1"><strong>错误信息</strong></div><div class="card-body py-2"><pre class="code-block code-block-error mb-0">' + esc(g.error) + '</pre></div></div>';
      }

      h += '</div>';
      document.getElementById('goroutineDetailContent').innerHTML = h;
      new bootstrap.Modal(document.getElementById('goroutineDetailModal')).show();
    },
    _formatName: function (g) {
      var st = g.stack || g.stack_trace;
      if (st) {
        var lines = st.split('\n').filter(function (l) { return l.trim(); });
        for (var i = 1; i < lines.length && i < 3; i += 2) {
          var fl = lines[i] ? lines[i].trim() : '';
          if (fl && fl.indexOf('goroutine ') === -1) return fl;
        }
      }
      if (g.name && g.name !== 'goroutine-' + g.id.replace(/^runtime-/, '')) return g.name;
      return 'Goroutine #' + g.id.replace(/^runtime-/, '');
    }
  };

  // ===========================
  // RequestModal
  // ===========================
  var RequestModal = {
    show: function (req) {
      var r = req;
      var reqHeaders = tryParseJson(r.req_header || (r.req_headers ? JSON.stringify(r.req_headers) : ''));
      var reqBody = tryParseJson(r.req_body);
      var respHeaders = tryParseJson(r.resp_header || (r.resp_headers ? JSON.stringify(r.resp_headers) : ''));
      var respBody = tryParseJson(r.resp_body);
      var statusCode = parseInt(r.resp_status_code) || r.status;

      var h = '<div class="modal-header"><h5 class="modal-title">请求详细信息</h5><button type="button" class="btn-close" data-bs-dismiss="modal"></button></div>';
      h += '<div class="modal-body">';
      h += '<table class="table table-sm table-bordered mb-3"><tbody>';
      h += '<tr><th style="width:100px">请求ID</th><td>' + esc(r.request_id) + '</td><th style="width:100px">状态</th><td><span class="badge ' + Fmt.statusBadge(statusCode) + '">' + esc(statusCode) + '</span></td></tr>';
      h += '<tr><th>方法</th><td><span class="badge bg-' + Fmt.methodBadge(r.req_method) + ' font-mono">' + Fmt.method(r.req_method) + '</span></td><th>URL</th><td class="font-mono small">' + esc(r.req_url) + '</td></tr>';
      h += '<tr><th>客户端IP</th><td class="font-mono">' + esc(r.ip) + '</td><th>User Agent</th><td class="small">' + esc(r.user_agent || 'N/A') + '</td></tr>';
      h += '<tr><th>延迟</th><td>' + Fmt.latency(r.latency) + '</td><th>开始时间</th><td>' + Fmt.dateTime(r.start_time) + '</td></tr>';
      if (r.end_time) h += '<tr><th>结束时间</th><td colspan="3">' + Fmt.dateTime(r.end_time) + '</td></tr>';
      h += '</tbody></table>';

      if (reqHeaders) h += '<div class="card mb-2"><div class="card-header py-1"><strong>请求头</strong></div><div class="card-body py-2">' + renderJsonOrPre(reqHeaders, r.req_header) + '</div></div>';
      if (r.req_body) h += '<div class="card mb-2"><div class="card-header py-1"><strong>请求体</strong></div><div class="card-body py-2">' + renderJsonOrPre(reqBody, r.req_body) + '</div></div>';
      if (respHeaders) h += '<div class="card mb-2"><div class="card-header py-1"><strong>响应头</strong></div><div class="card-body py-2">' + renderJsonOrPre(respHeaders, r.resp_header) + '</div></div>';
      if (r.resp_body) h += '<div class="card mb-2"><div class="card-header py-1"><strong>响应体</strong></div><div class="card-body py-2">' + renderJsonOrPre(respBody, r.resp_body) + '</div></div>';

      var sessionStr = r.session_logs || '';
      if (sessionStr && sessionStr.trim() !== '' && sessionStr !== '[]') {
        h += '<div class="card mb-2"><div class="card-header py-1"><strong>请求日志</strong></div><div class="card-body py-2">' + SessionLogs.render(sessionStr, false) + '</div></div>';
      }

      if (r.call_stack) h += '<div class="card mb-2"><div class="card-header py-1"><strong>调用栈</strong></div><div class="card-body py-2"><pre class="code-block code-block-info mb-0">' + esc(r.call_stack) + '</pre></div></div>';
      if (r.error) h += '<div class="card mb-2"><div class="card-header py-1"><strong>错误信息</strong></div><div class="card-body py-2"><pre class="code-block code-block-error mb-0">' + esc(r.error) + '</pre></div></div>';

      h += '</div>';
      document.getElementById('requestDetailContent').innerHTML = h;
      new bootstrap.Modal(document.getElementById('requestDetailModal')).show();
    }
  };

  // ===========================
  // BatchLogsModal
  // ===========================
  var BatchLogsModal = {
    show: function (requests, sortNewerFirst) {
      if (sortNewerFirst === undefined) sortNewerFirst = true;
      var filtered = requests.filter(function (r) {
        return r.session_logs && r.session_logs.trim() !== '' && r.session_logs !== '[]';
      });
      filtered.sort(function (a, b) { return sortNewerFirst ? b.start_time - a.start_time : a.start_time - b.start_time; });

      var h = '<div class="modal-header"><h5 class="modal-title">批量请求日志查看</h5><button type="button" class="btn-close" data-bs-dismiss="modal"></button></div>';
      h += '<div class="modal-body">';
      h += '<div class="alert alert-info py-2"><i class="bi bi-info-circle me-1"></i>总计: ' + requests.length + ' 条记录，包含会话日志: ' + filtered.length + ' 条记录</div>';

      if (filtered.length > 1) {
        h += '<div class="sort-control"><span class="fw-medium text-muted me-2">排序顺序:</span>' +
          '<div class="btn-group btn-group-sm"><button class="btn btn-outline-primary batch-sort-btn active" data-order="newer">新的在前</button>' +
          '<button class="btn btn-outline-primary batch-sort-btn" data-order="older">旧的在前</button></div></div>';
      }

      if (filtered.length === 0) {
        h += '<div class="empty-state"><i class="bi bi-inbox"></i><p>选中的记录中没有可用的会话日志</p></div>';
      } else {
        h += '<div class="accordion" id="batchAccordion">';
        filtered.forEach(function (req, idx) {
          var collapseId = 'batchCollapse' + idx;
          h += '<div class="accordion-item"><h2 class="accordion-header"><button class="accordion-button" type="button" data-bs-toggle="collapse" data-bs-target="#' + collapseId + '">' +
            '<div class="session-header"><span class="badge bg-' + Fmt.methodBadge(req.req_method) + ' fw-bold">' + esc(req.req_method) + '</span>' +
            '<span class="url-text">' + esc(req.req_url) + '</span>' +
            '<span class="badge bg-info text-dark font-mono">' + esc(req.latency) + '</span>' +
            '<span class="time-text">' + new Date(req.start_time * 1000).toLocaleString() + '</span></div>' +
            '</button></h2><div id="' + collapseId + '" class="accordion-collapse collapse show" data-bs-parent="#batchAccordion"><div class="accordion-body">' +
            '<table class="table table-sm table-bordered mb-3"><tbody>' +
            '<tr><th>请求时间</th><td>' + new Date(req.start_time * 1000).toLocaleString() + '</td><th>请求ID</th><td class="font-mono small">' + esc(req.request_id) + '</td></tr>' +
            '<tr><th>状态码</th><td>' + esc(req.resp_status_code) + '</td><th>客户端IP</th><td>' + esc(req.ip) + '</td></tr>' +
            (req.user_agent ? '<tr><th>User Agent</th><td colspan="3" class="small">' + esc(req.user_agent) + '</td></tr>' : '') +
            '</tbody></table>' +
            SessionLogs.render(req.session_logs, false) +
            '</div></div></div>';
        });
        h += '</div>';
      }
      h += '</div>';
      document.getElementById('batchLogsContent').innerHTML = h;
      new bootstrap.Modal(document.getElementById('batchLogsModal')).show();
    }
  };

  // ===========================
  // StackTraceModal
  // ===========================
  var StackTraceModal = {
    show: function (entry) {
      var e = entry;
      var h = '<div class="modal-header"><h5 class="modal-title">函数调用栈详情</h5><button type="button" class="btn-close" data-bs-dismiss="modal"></button></div>';
      h += '<div class="modal-body">';

      h += '<div class="card mb-2"><div class="card-header py-1"><strong>分配统计</strong></div><div class="card-body py-2">' +
        '<div class="row text-center">' +
        '<div class="col-6 col-md-3"><div class="fs-6 fw-semibold text-primary">' + e.inuse_objects.toLocaleString() + '</div><small class="text-muted">使用中对象</small></div>' +
        '<div class="col-6 col-md-3"><div class="fs-6 fw-semibold text-success">' + Fmt.bytes(e.inuse_bytes) + '</div><small class="text-muted">使用中内存</small></div>' +
        '<div class="col-6 col-md-3"><div class="fs-6 fw-semibold text-warning">' + e.alloc_objects.toLocaleString() + '</div><small class="text-muted">总分配对象</small></div>' +
        '<div class="col-6 col-md-3"><div class="fs-6 fw-semibold text-danger">' + Fmt.bytes(e.alloc_bytes) + '</div><small class="text-muted">总分配内存</small></div>' +
        '</div></div></div>';

      h += '<div class="card mb-2"><div class="card-header py-1"><strong>主要函数</strong></div><div class="card-body py-2"><code>' + esc(e.top_function) + '</code></div></div>';

      h += '<div class="card mb-2"><div class="card-header py-1"><strong>完整调用栈</strong> <span class="badge bg-primary ms-2">' + (e.stack_trace || []).length + ' 层</span></div><div class="card-body py-2">';
      if (e.stack_trace && e.stack_trace.length > 0) {
        e.stack_trace.forEach(function (func, idx) {
          var parts = func.split('\n');
          h += '<div class="stack-item"><div class="stack-number">' + (idx + 1) + '</div><div class="flex-grow-1">' +
            '<div class="stack-func">' + esc(parts[0]) + '</div>';
          if (parts[1]) h += '<div class="stack-file">' + esc(parts[1].trim()) + '</div>';
          if (idx === 0) h += '<small class="text-primary fw-medium mt-1 d-block">← 分配点</small>';
          h += '</div></div>';
        });
      } else {
        h += '<div class="empty-state"><i class="bi bi-code-slash"></i><p>暂无调用栈信息</p></div>';
      }
      h += '</div></div>';
      h += '</div>';

      document.getElementById('stackTraceContent').innerHTML = h;
      new bootstrap.Modal(document.getElementById('stackTraceModal')).show();
    }
  };

  // ===========================
  // Pagination helper
  // ===========================
  function renderPagination(total, page, pageSize, onPageChange) {
    var totalPages = Math.ceil(total / pageSize);
    if (totalPages <= 1 && total <= pageSize) return '';
    var start = (page - 1) * pageSize + 1;
    var end = Math.min(page * pageSize, total);

    var h = '<div class="pagination-wrapper">';
    h += '<div class="pagination-info">第 ' + start + '-' + end + ' 条，共 ' + total + ' 条</div>';
    h += '<div class="d-flex align-items-center gap-2">';
    h += '<select class="form-select form-select-sm page-size-select" style="width:auto">';
    [50, 100, 200, 500].forEach(function (s) {
      h += '<option value="' + s + '"' + (s === pageSize ? ' selected' : '') + '>' + s + ' 条/页</option>';
    });
    h += '</select>';
    h += '<nav><ul class="pagination pagination-sm mb-0">';
    h += '<li class="page-item' + (page <= 1 ? ' disabled' : '') + '"><a class="page-link" href="#" data-page="' + (page - 1) + '">‹</a></li>';

    var pages = [];
    if (totalPages <= 7) {
      for (var i = 1; i <= totalPages; i++) pages.push(i);
    } else {
      pages = [1];
      if (page > 3) pages.push('...');
      for (var j = Math.max(2, page - 1); j <= Math.min(totalPages - 1, page + 1); j++) pages.push(j);
      if (page < totalPages - 2) pages.push('...');
      pages.push(totalPages);
    }
    pages.forEach(function (p) {
      if (p === '...') { h += '<li class="page-item disabled"><span class="page-link">…</span></li>'; }
      else { h += '<li class="page-item' + (p === page ? ' active' : '') + '"><a class="page-link" href="#" data-page="' + p + '">' + p + '</a></li>'; }
    });
    h += '<li class="page-item' + (page >= totalPages ? ' disabled' : '') + '"><a class="page-link" href="#" data-page="' + (page + 1) + '">›</a></li>';
    h += '</ul></nav></div></div>';
    return h;
  }

  // ===========================
  // Sortable table helper
  // ===========================
  function sortData(data, sortCol, sortDir) {
    if (!sortCol) return data;
    return data.slice().sort(function (a, b) {
      var va = a[sortCol], vb = b[sortCol];
      if (va === undefined) va = 0;
      if (vb === undefined) vb = 0;
      if (typeof va === 'string') { va = va.toLowerCase(); vb = (vb || '').toLowerCase(); }
      var cmp = va < vb ? -1 : (va > vb ? 1 : 0);
      return sortDir === 'desc' ? -cmp : cmp;
    });
  }

  // ===========================
  // Parse stack trace helper
  // ===========================
  function parseStackTrace(stack) {
    if (!stack) return [];
    var stacks = [];
    if (typeof stack === 'string') {
      var lines = stack.split('\n').filter(function (l) { return l.trim(); });
      for (var i = 1; i < lines.length && stacks.length < 2; i += 2) {
        var funcLine = lines[i] ? lines[i].trim() : '';
        var fileLine = lines[i + 1] ? lines[i + 1].trim() : '';
        if (funcLine && fileLine) {
          var fm = fileLine.match(/^(.+\.go):(\d+)\s+/);
          stacks.push({ func: funcLine, file: fm ? fm[1] : fileLine, line: fm ? fm[2] : '' });
        }
      }
    } else if (Array.isArray(stack)) {
      for (var j = 0; j < Math.min(stack.length, 2); j++) {
        var entry = stack[j];
        if (!entry) continue;
        if (entry.indexOf('\n') >= 0) {
          var parts = entry.split('\n');
          var fn = parts[0] ? parts[0].trim() : '';
          var fl = parts[1] ? parts[1].trim() : '';
          var fm2 = fl.match(/^(.+\.go):(\d+)$/);
          stacks.push({ func: fn, file: fm2 ? fm2[1] : fl, line: fm2 ? fm2[2] : '' });
        } else {
          stacks.push({ func: entry, file: '', line: '' });
        }
      }
    }
    return stacks;
  }

  // ===========================
  // Dashboard Page
  // ===========================
  var DashboardPage = {
    interval: null,
    render: function () {
      var el = document.getElementById('app-content');
      el.innerHTML = '<h2 class="page-title"><i class="bi bi-speedometer2"></i>仪表盘</h2>' +
        '<div class="row g-2 mb-3" id="dash-stats"></div>' +
        '<div class="row g-2"><div class="col-lg-6"><div class="card" id="dash-goroutines"></div></div><div class="col-lg-6"><div class="card" id="dash-requests"></div></div></div>';
      this._load();
      var self = this;
      WS.onChange(function () { self._updateStats(); });
      this.interval = setInterval(function () { if (!WS.connected) self._load(); }, 30000);
    },
    destroy: function () { if (this.interval) { clearInterval(this.interval); this.interval = null; } },
    _data: {},
    _load: function () {
      var self = this;
      Promise.all([Api.getSystemInfo(), Api.getGoroutines({ limit: 5 }), Api.getRequests({ limit: 5 }), Api.getHeapProfile()])
        .then(function (res) { self._data = { system: res[0], goroutines: res[1], requests: res[2], heap: res[3] }; self._updateStats(); self._updateLists(); })
        .catch(function (e) { console.error('Dashboard load error:', e); });
    },
    _updateStats: function () {
      var ws = WS.systemStats, d = this._data;
      var activeG = (ws.goroutine_stats && ws.goroutine_stats.active_count) || (d.goroutines && d.goroutines.total) || (d.system && d.system.goroutines && d.system.goroutines.total) || 0;
      var totalReq = (ws.request_stats && ws.request_stats.total_requests) || (d.requests && d.requests.total) || 0;
      var heapObj = (d.heap && d.heap.total_inuse_objects) || 0;
      var mem = (ws.system_stats && ws.system_stats.memory_usage) || (d.system && d.system.memory && d.system.memory.heap_alloc) || 0;
      document.getElementById('dash-stats').innerHTML =
        this._statCard('bi-lightning-charge text-primary', '活跃协程', Fmt.number(activeG)) +
        this._statCard('bi-bar-chart text-success', '总请求数', Fmt.number(totalReq)) +
        this._statCard('bi-database text-warning', 'Heap', Fmt.number(heapObj)) +
        this._statCard('bi-memory text-danger', '内存使用', Fmt.bytes(mem));
    },
    _statCard: function (icon, label, value) {
      return '<div class="col-sm-6 col-lg-3"><div class="card stat-card"><div class="card-body d-flex align-items-center gap-3">' +
        '<i class="bi ' + icon + ' stat-icon"></i><div><div class="stat-value">' + esc(value) + '</div><div class="stat-label">' + esc(label) + '</div></div></div></div></div>';
    },
    _updateLists: function () {
      var d = this._data;
      var goroutines = WS.recentGoroutines.length > 0 ? WS.recentGoroutines.slice(0, 5) : this._mapGoroutines(d.goroutines);
      var requests = WS.recentRequests.length > 0 ? WS.recentRequests.slice(0, 5) : this._mapRequests(d.requests);

      var gh = '<div class="card-header d-flex justify-content-between align-items-center"><span><i class="bi bi-lightning-charge me-1"></i>最近协程活动</span><a href="#/goroutines" class="text-decoration-none small">查看全部 →</a></div>';
      gh += '<div class="card-body p-0">';
      if (goroutines.length === 0) {
        var totalG = d.goroutines && d.goroutines.total ? d.goroutines.total : 0;
        if (totalG > 0) {
          gh += '<div class="text-center py-4"><i class="bi bi-info-circle text-muted fs-3 d-block mb-2"></i><p class="text-muted mb-1">协程跟踪未启用</p><small class="text-muted">系统当前有 ' + totalG + ' 个协程运行</small></div>';
        } else {
          gh += '<div class="text-center py-4 text-muted">暂无协程数据</div>';
        }
      } else {
        gh += '<ul class="list-group list-group-flush">';
        goroutines.forEach(function (g) {
          var name = GoroutineModal._formatName(g);
          var dur = g.duration ? Fmt.duration(g.duration * 1000) : Fmt.duration((Math.floor(Date.now() / 1000) - g.start_time) * 1000);
          gh += '<li class="list-group-item"><div class="d-flex justify-content-between align-items-start">' +
            '<div class="flex-grow-1 min-width-0">' +
            '<div class="d-flex align-items-center gap-2 mb-1"><span class="badge ' + Fmt.statusBadge(g.status) + '">' + esc(g.status) + '</span>' +
            '<span class="font-mono small text-truncate" style="max-width:70%">' + esc(name) + '</span></div>' +
            '<small class="text-muted">运行时长: ' + dur + ' | 开始: ' + Fmt.time(g.start_time) + '</small>' +
            '</div><button class="btn btn-sm btn-link dash-goroutine-detail" data-idx="' + goroutines.indexOf(g) + '">查看</button></div></li>';
        });
        gh += '</ul>';
      }
      gh += '</div>';
      document.getElementById('dash-goroutines').innerHTML = gh;

      var rh = '<div class="card-header d-flex justify-content-between align-items-center"><span><i class="bi bi-bar-chart me-1"></i>最近请求记录</span><a href="#/requests" class="text-decoration-none small">查看全部 →</a></div>';
      rh += '<div class="card-body p-0">';
      if (requests.length === 0) {
        rh += '<div class="text-center py-4 text-muted">暂无请求数据</div>';
      } else {
        rh += '<ul class="list-group list-group-flush">';
        requests.forEach(function (r) {
          rh += '<li class="list-group-item"><div class="d-flex justify-content-between align-items-start">' +
            '<div class="flex-grow-1 min-width-0">' +
            '<div class="d-flex align-items-center gap-2 mb-1">' +
            '<span class="badge ' + Fmt.statusBadge(r.resp_status_code) + '">' + (r.resp_status_code || r.status) + '</span>' +
            '<span class="badge bg-' + Fmt.methodBadge(r.req_method) + '">' + esc(r.req_method) + '</span>' +
            '<span class="small text-truncate" style="max-width:60%">' + esc(r.req_url) + '</span></div>' +
            '<small class="text-muted">IP: ' + esc(r.ip) + (r.latency ? ' | 延迟: ' + Fmt.latency(r.latency) : '') + ' | 时间: ' + Fmt.time(r.start_time) + '</small>' +
            '</div><button class="btn btn-sm btn-link dash-request-detail" data-id="' + esc(r.request_id) + '">查看</button></div></li>';
        });
        rh += '</ul>';
      }
      rh += '</div>';
      document.getElementById('dash-requests').innerHTML = rh;

      this._bindEvents(goroutines, requests);
    },
    _mapGoroutines: function (data) {
      if (!data || !data.data) return [];
      return data.data.slice(0, 5).map(function (t) {
        return { id: t.id, name: t.name, status: t.status, duration: t.end_time ? t.end_time - t.start_time : Math.floor(Date.now() / 1000) - t.start_time, start_time: t.start_time, stack_trace: t.stack, stack: t.stack };
      });
    },
    _mapRequests: function (data) {
      if (!data || !data.data) return [];
      return data.data.slice(0, 5).map(function (t) {
        return { request_id: t.request_id, req_method: t.req_method, req_url: t.req_url, resp_status_code: parseInt(t.resp_status_code) || 0, status: t.status || 'completed', ip: t.ip, latency: t.latency, start_time: t.start_time };
      });
    },
    _bindEvents: function (goroutines, requests) {
      document.querySelectorAll('.dash-goroutine-detail').forEach(function (btn) {
        btn.addEventListener('click', function () {
          var g = goroutines[parseInt(btn.dataset.idx)];
          if (g) GoroutineModal.show(g);
        });
      });
      document.querySelectorAll('.dash-request-detail').forEach(function (btn) {
        btn.addEventListener('click', function () {
          var id = btn.dataset.id;
          Api.getRequestDetail(id).then(function (detail) { RequestModal.show(detail); }).catch(function () {
            var r = requests.find(function (r) { return r.request_id === id; });
            if (r) RequestModal.show(r);
          });
        });
      });
    }
  };

  // ===========================
  // Goroutines Page
  // ===========================
  var GoroutinesPage = {
    interval: null, state: { search: '', status: '', type: '', page: 1, pageSize: 50, sortCol: '', sortDir: 'asc' }, data: null,
    render: function () {
      document.getElementById('app-content').innerHTML = '<h2 class="page-title"><i class="bi bi-lightning-charge"></i>协程监控</h2><div id="goroutines-controls"></div><div class="card"><div class="card-body p-0" id="goroutines-table"></div></div>';
      this._renderControls();
      this._load();
      var self = this;
      this.interval = setInterval(function () { if (!WS.connected) self._load(); }, 30000);
    },
    destroy: function () { if (this.interval) { clearInterval(this.interval); this.interval = null; } },
    _renderControls: function () {
      var s = this.state;
      document.getElementById('goroutines-controls').innerHTML =
        '<div class="controls-wrapper">' +
        '<div class="input-group" style="width:250px"><span class="input-group-text"><i class="bi bi-search"></i></span><input type="text" class="form-control" id="g-search" placeholder="搜索协程名称或 ID" value="' + esc(s.search) + '"></div>' +
        '<select class="form-select" id="g-status" style="width:150px"><option value="">筛选状态</option><option value="running">运行中</option><option value="waiting">等待中</option><option value="completed">已完成</option><option value="failed">已失败</option><option value="blocked">阻塞中</option></select>' +
        '<select class="form-select" id="g-type" style="width:120px"><option value="">协程类型</option><option value="active">活跃协程</option><option value="history">历史记录</option></select>' +
        '<button class="btn btn-primary" id="g-refresh"><i class="bi bi-arrow-clockwise me-1"></i>刷新数据</button>' +
        '</div>';
      var self = this;
      document.getElementById('g-search').addEventListener('input', function (e) { self.state.search = e.target.value; self.state.page = 1; self._renderTable(); });
      document.getElementById('g-status').addEventListener('change', function (e) { self.state.status = e.target.value; self.state.page = 1; self._renderTable(); });
      document.getElementById('g-type').addEventListener('change', function (e) { self.state.type = e.target.value; self._load(); });
      document.getElementById('g-refresh').addEventListener('click', function () { self._load(); });
    },
    _load: function () {
      var self = this, params = {};
      if (self.state.type) params.type = self.state.type;
      Api.getGoroutines(params).then(function (d) { self.data = d; self._renderTable(); }).catch(function (e) { console.error(e); });
    },
    _getFiltered: function () {
      if (!this.data || !this.data.data) return [];
      var s = this.state;
      return this.data.data.filter(function (item) {
        var matchSearch = !s.search || (item.name || '').toLowerCase().indexOf(s.search.toLowerCase()) >= 0 || item.id.indexOf(s.search) >= 0;
        var matchStatus = !s.status || item.status === s.status;
        return matchSearch && matchStatus;
      });
    },
    _renderTable: function () {
      var filtered = this._getFiltered();
      var s = this.state;
      if (s.sortCol) {
        filtered = filtered.slice().sort(function (a, b) {
          var va, vb;
          if (s.sortCol === 'duration') {
            var now = Math.floor(Date.now() / 1000);
            va = a.end_time ? a.end_time - a.start_time : now - a.start_time;
            vb = b.end_time ? b.end_time - b.start_time : now - b.start_time;
          } else {
            va = a[s.sortCol] || 0; vb = b[s.sortCol] || 0;
          }
          return s.sortDir === 'desc' ? vb - va : va - vb;
        });
      }
      var total = filtered.length;
      var start = (s.page - 1) * s.pageSize;
      var pageData = filtered.slice(start, start + s.pageSize);

      var h = '<div class="table-responsive"><table class="table table-sm table-striped table-hover mb-0"><thead><tr>' +
        '<th style="width:100px">ID</th><th style="width:400px">名称/函数</th><th style="width:100px">状态</th><th style="width:60px">日志</th>' +
        '<th class="th-sortable' + (s.sortCol === 'duration' ? ' sort-' + s.sortDir : '') + '" data-sort="duration" style="width:90px">运行时长 <i class="bi bi-arrow-down-up sort-icon"></i></th>' +
        '<th class="th-sortable' + (s.sortCol === 'start_time' ? ' sort-' + s.sortDir : '') + '" data-sort="start_time" style="width:120px">开始时间 <i class="bi bi-arrow-down-up sort-icon"></i></th>' +
        '<th style="width:80px">操作</th></tr></thead><tbody>';

      if (pageData.length === 0) {
        h += '<tr><td colspan="7"><div class="empty-state"><i class="bi bi-lightning-charge"></i><p>暂无协程数据</p></div></td></tr>';
      } else {
        var now = Math.floor(Date.now() / 1000);
        pageData.forEach(function (g) {
          var dur = g.end_time ? g.end_time - g.start_time : now - g.start_time;
          var stacks = parseStackTrace(g.stack);
          var nameHtml = '<div class="func-cell"><div class="func-name">';
          if (g.name && g.name !== 'goroutine-' + g.id.replace(/^runtime-/, '')) nameHtml += esc(g.name);
          else nameHtml += '<span class="text-muted">Goroutine #' + esc(g.id.replace(/^runtime-/, '')) + '</span>';
          nameHtml += '</div>';
          if (stacks.length > 0) {
            nameHtml += '<div class="stack-detail">';
            stacks.forEach(function (st) { nameHtml += '<div><span class="text-primary">' + esc(st.func) + '</span><br><span class="text-muted">' + esc(st.file) + ':' + esc(st.line) + '</span></div>'; });
            nameHtml += '</div>';
          }
          nameHtml += '</div>';

          h += '<tr><td class="font-mono small">' + esc(g.id.replace(/^runtime-/, '')) + '</td><td>' + nameHtml + '</td>' +
            '<td><span class="badge ' + Fmt.statusBadge(g.status) + '">' + esc(g.status) + '</span></td>' +
            '<td>' + (g.session_logs && g.session_logs.length > 0 ? '<span class="text-success fw-bold">✓</span>' : '<span class="text-muted">-</span>') + '</td>' +
            '<td>' + Fmt.duration(dur * 1000) + '</td><td>' + Fmt.time(g.start_time) + '</td>' +
            '<td><button class="btn btn-sm btn-link goroutine-view" data-id="' + esc(g.id) + '">查看</button></td></tr>';
        });
      }
      h += '</tbody></table></div>';
      h += renderPagination(total, s.page, s.pageSize);

      document.getElementById('goroutines-table').innerHTML = h;
      this._bindTableEvents(pageData);
    },
    _bindTableEvents: function (pageData) {
      var self = this;
      document.querySelectorAll('#goroutines-table .goroutine-view').forEach(function (btn) {
        btn.addEventListener('click', function () {
          var id = btn.dataset.id;
          Api.getGoroutineDetail(id).then(function (d) { GoroutineModal.show(d); }).catch(function () {
            var g = pageData.find(function (g) { return g.id === id; });
            if (g) GoroutineModal.show(g);
          });
        });
      });
      document.querySelectorAll('#goroutines-table .th-sortable').forEach(function (th) {
        th.addEventListener('click', function () {
          var col = th.dataset.sort;
          if (self.state.sortCol === col) self.state.sortDir = self.state.sortDir === 'asc' ? 'desc' : 'asc';
          else { self.state.sortCol = col; self.state.sortDir = 'asc'; }
          self._renderTable();
        });
      });
      document.querySelectorAll('#goroutines-table .page-link[data-page]').forEach(function (a) {
        a.addEventListener('click', function (e) { e.preventDefault(); self.state.page = parseInt(a.dataset.page); self._renderTable(); });
      });
      var pss = document.querySelector('#goroutines-table .page-size-select');
      if (pss) pss.addEventListener('change', function (e) { self.state.pageSize = parseInt(e.target.value); self.state.page = 1; self._renderTable(); });
    }
  };

  // ===========================
  // Requests Page
  // ===========================
  var RequestsPage = {
    interval: null, state: { search: '', method: '', status: '', page: 1, pageSize: 50, sortCol: '', sortDir: 'asc', selected: {} }, data: null,
    render: function () {
      document.getElementById('app-content').innerHTML = '<h2 class="page-title"><i class="bi bi-globe"></i>请求监控</h2><div id="requests-controls"></div><div id="requests-selection-alert"></div><div class="card"><div class="card-body p-0" id="requests-table"></div></div>';
      this._renderControls();
      this._load();
      var self = this;
      this.interval = setInterval(function () { if (!WS.connected) self._load(); }, 30000);
    },
    destroy: function () { if (this.interval) { clearInterval(this.interval); this.interval = null; } },
    _renderControls: function () {
      var s = this.state;
      document.getElementById('requests-controls').innerHTML =
        '<div class="controls-wrapper">' +
        '<div class="input-group" style="width:250px"><span class="input-group-text"><i class="bi bi-search"></i></span><input type="text" class="form-control" id="r-search" placeholder="搜索路径、IP或ID" value="' + esc(s.search) + '"></div>' +
        '<select class="form-select" id="r-method" style="width:120px"><option value="">筛选方法</option><option>GET</option><option>POST</option><option>PUT</option><option>DELETE</option><option>PATCH</option></select>' +
        '<select class="form-select" id="r-status" style="width:150px"><option value="">筛选状态码</option><option value="200">200 - 成功</option><option value="400">400 - 错误请求</option><option value="401">401 - 未授权</option><option value="404">404 - 未找到</option><option value="500">500 - 服务器错误</option></select>' +
        '<button class="btn btn-primary" id="r-refresh"><i class="bi bi-arrow-clockwise me-1"></i>刷新数据</button>' +
        '<span class="vr"></span>' +
        '<button class="btn btn-outline-secondary" id="r-select-all">全选</button>' +
        '<button class="btn btn-outline-secondary" id="r-clear-sel">清除选择</button>' +
        '<button class="btn btn-primary" id="r-batch-logs" disabled><i class="bi bi-eye me-1"></i>批量查看日志 (<span id="r-sel-count">0</span>)</button>' +
        '</div>';
      var self = this;
      document.getElementById('r-search').addEventListener('input', function (e) { self.state.search = e.target.value; self.state.page = 1; self._renderTable(); });
      document.getElementById('r-method').addEventListener('change', function (e) { self.state.method = e.target.value; self.state.page = 1; self._renderTable(); });
      document.getElementById('r-status').addEventListener('change', function (e) { self.state.status = e.target.value; self.state.page = 1; self._renderTable(); });
      document.getElementById('r-refresh').addEventListener('click', function () { self.state.selected = {}; self._load(); });
      document.getElementById('r-select-all').addEventListener('click', function () {
        self._getFiltered().forEach(function (r) { self.state.selected[r.request_id] = true; });
        self._renderTable(); self._updateSelectionUI();
      });
      document.getElementById('r-clear-sel').addEventListener('click', function () { self.state.selected = {}; self._renderTable(); self._updateSelectionUI(); });
      document.getElementById('r-batch-logs').addEventListener('click', function () { self._showBatchLogs(); });
    },
    _load: function () {
      var self = this;
      Api.getRequests().then(function (d) { self.data = d; self._renderTable(); }).catch(function (e) { console.error(e); });
    },
    _getFiltered: function () {
      if (!this.data || !this.data.data) return [];
      var s = this.state;
      return this.data.data.filter(function (item) {
        var matchSearch = !s.search || (item.req_url || '').toLowerCase().indexOf(s.search.toLowerCase()) >= 0 || (item.ip || '').indexOf(s.search) >= 0 || (item.request_id || '').indexOf(s.search) >= 0;
        var matchMethod = !s.method || item.req_method === s.method;
        var matchStatus = !s.status || String(item.resp_status_code) === s.status;
        return matchSearch && matchMethod && matchStatus;
      });
    },
    _updateSelectionUI: function () {
      var count = Object.keys(this.state.selected).length;
      var countEl = document.getElementById('r-sel-count');
      if (countEl) countEl.textContent = count;
      var batchBtn = document.getElementById('r-batch-logs');
      if (batchBtn) batchBtn.disabled = count === 0;
      var alertEl = document.getElementById('requests-selection-alert');
      if (alertEl) {
        alertEl.innerHTML = count > 0 ? '<div class="alert alert-info py-2 mb-3"><i class="bi bi-info-circle me-1"></i>已选择 ' + count + ' 条记录</div>' : '';
      }
    },
    _renderTable: function () {
      var filtered = this._getFiltered();
      var s = this.state;
      if (s.sortCol) {
        filtered = filtered.slice().sort(function (a, b) {
          var va = a[s.sortCol], vb = b[s.sortCol];
          if (s.sortCol === 'resp_status_code') { va = parseInt(va) || 0; vb = parseInt(vb) || 0; }
          if (typeof va === 'string') { va = va.toLowerCase(); vb = (vb || '').toLowerCase(); }
          return s.sortDir === 'desc' ? (vb > va ? 1 : -1) : (va > vb ? 1 : -1);
        });
      }
      var total = filtered.length;
      var start = (s.page - 1) * s.pageSize;
      var pageData = filtered.slice(start, start + s.pageSize);

      var h = '<div class="table-responsive"><table class="table table-sm table-striped table-hover mb-0"><thead><tr>' +
        '<th style="width:40px"><input type="checkbox" class="form-check-input" id="r-check-all"></th>' +
        '<th style="width:80px">方法</th><th>请求路径</th>' +
        '<th class="th-sortable' + (s.sortCol === 'resp_status_code' ? ' sort-' + s.sortDir : '') + '" data-sort="resp_status_code" style="width:100px">状态码 <i class="bi bi-arrow-down-up sort-icon"></i></th>' +
        '<th style="width:120px">IP地址</th><th style="width:100px">延迟</th>' +
        '<th class="th-sortable' + (s.sortCol === 'start_time' ? ' sort-' + s.sortDir : '') + '" data-sort="start_time" style="width:150px">开始时间 <i class="bi bi-arrow-down-up sort-icon"></i></th>' +
        '<th style="width:80px">操作</th></tr></thead><tbody>';

      if (pageData.length === 0) {
        h += '<tr><td colspan="8"><div class="empty-state"><i class="bi bi-globe"></i><p>暂无请求数据</p></div></td></tr>';
      } else {
        pageData.forEach(function (r) {
          var checked = s.selected[r.request_id] ? ' checked' : '';
          var sc = parseInt(r.resp_status_code) || r.status;
          h += '<tr><td><input type="checkbox" class="form-check-input r-row-check" data-id="' + esc(r.request_id) + '"' + checked + '></td>' +
            '<td><span class="badge bg-' + Fmt.methodBadge(r.req_method) + ' font-mono">' + Fmt.method(r.req_method) + '</span></td>' +
            '<td class="font-mono small text-truncate" style="max-width:300px" title="' + esc(r.req_url) + '">' + esc(r.req_url) + '</td>' +
            '<td><span class="badge ' + Fmt.statusBadge(sc) + '">' + esc(sc) + '</span></td>' +
            '<td class="font-mono small">' + esc(r.ip) + '</td>' +
            '<td>' + Fmt.latency(r.latency) + '</td><td>' + Fmt.time(r.start_time) + '</td>' +
            '<td><button class="btn btn-sm btn-link request-view" data-id="' + esc(r.request_id) + '">查看</button></td></tr>';
        });
      }
      h += '</tbody></table></div>';
      h += renderPagination(total, s.page, s.pageSize);

      document.getElementById('requests-table').innerHTML = h;
      this._bindTableEvents(pageData);
      this._updateSelectionUI();
    },
    _bindTableEvents: function (pageData) {
      var self = this;
      document.querySelectorAll('#requests-table .request-view').forEach(function (btn) {
        btn.addEventListener('click', function () {
          Api.getRequestDetail(btn.dataset.id).then(function (d) { RequestModal.show(d); }).catch(function () {
            var r = pageData.find(function (r) { return r.request_id === btn.dataset.id; });
            if (r) RequestModal.show(r);
          });
        });
      });
      document.querySelectorAll('#requests-table .r-row-check').forEach(function (cb) {
        cb.addEventListener('change', function () {
          if (cb.checked) self.state.selected[cb.dataset.id] = true;
          else delete self.state.selected[cb.dataset.id];
          self._updateSelectionUI();
        });
      });
      var checkAll = document.getElementById('r-check-all');
      if (checkAll) {
        checkAll.addEventListener('change', function () {
          pageData.forEach(function (r) {
            if (checkAll.checked) self.state.selected[r.request_id] = true;
            else delete self.state.selected[r.request_id];
          });
          self._renderTable();
        });
      }
      document.querySelectorAll('#requests-table .th-sortable').forEach(function (th) {
        th.addEventListener('click', function () {
          var col = th.dataset.sort;
          if (self.state.sortCol === col) self.state.sortDir = self.state.sortDir === 'asc' ? 'desc' : 'asc';
          else { self.state.sortCol = col; self.state.sortDir = 'asc'; }
          self._renderTable();
        });
      });
      document.querySelectorAll('#requests-table .page-link[data-page]').forEach(function (a) {
        a.addEventListener('click', function (e) { e.preventDefault(); self.state.page = parseInt(a.dataset.page); self._renderTable(); });
      });
      var pss = document.querySelector('#requests-table .page-size-select');
      if (pss) pss.addEventListener('change', function (e) { self.state.pageSize = parseInt(e.target.value); self.state.page = 1; self._renderTable(); });
    },
    _showBatchLogs: function () {
      var self = this;
      var ids = Object.keys(self.state.selected);
      if (ids.length === 0) return;
      Promise.all(ids.map(function (id) { return Api.getRequestDetail(id); }))
        .then(function (details) { BatchLogsModal.show(details); })
        .catch(function (e) { console.error('Batch logs error:', e); });
    }
  };

  // ===========================
  // System Page
  // ===========================
  var SystemPage = {
    interval: null, data: null,
    render: function () {
      document.getElementById('app-content').innerHTML = '<h2 class="page-title"><i class="bi bi-pc-display"></i>系统信息</h2><div class="row g-2 mb-3" id="sys-stats"></div><div class="row g-2"><div class="col-lg-6"><div class="card" id="sys-info"></div></div><div class="col-lg-6"><div class="card" id="sys-runtime"></div></div></div>';
      this._load();
      var self = this;
      WS.onChange(function () { self._update(); });
      this.interval = setInterval(function () { if (!WS.connected) self._load(); }, 30000);
    },
    destroy: function () { if (this.interval) { clearInterval(this.interval); this.interval = null; } },
    _load: function () {
      var self = this;
      Api.getSystemInfo().then(function (d) { self.data = d; self._update(); }).catch(function (e) { console.error(e); });
    },
    _update: function () {
      var ws = WS.systemStats, d = this.data || {};
      var mem = (ws.system_stats && ws.system_stats.memory_usage) || (d.memory && d.memory.alloc) || 0;
      var cpu = (ws.system_stats && ws.system_stats.cpu_usage) || 0;
      var activeG = (ws.goroutine_stats && ws.goroutine_stats.active_count) || (d.goroutines && d.goroutines.total) || 0;
      var totalG = (ws.goroutine_stats && ws.goroutine_stats.total_count) || (d.goroutines && d.goroutines.total) || 0;
      var uptime = (ws.system_stats && ws.system_stats.uptime) || (d.startup_time && d.timestamp ? d.timestamp - d.startup_time : 0);
      var totalReq = (ws.request_stats && ws.request_stats.total_requests) || 0;
      var successRate = (ws.request_stats && ws.request_stats.success_rate) || 0;
      var memPct = Math.min((mem / (1024 * 1024 * 1024 * 8)) * 100, 100).toFixed(1);

      document.getElementById('sys-stats').innerHTML =
        this._statCard('bi-database text-purple', '内存使用', Fmt.bytes(mem), '<div class="progress mt-1" style="height:4px"><div class="progress-bar bg-purple" style="width:' + memPct + '%"></div></div>') +
        this._statCard('bi-lightning-charge text-warning', 'CPU使用率', cpu.toFixed(2) + '%', '<div class="progress mt-1" style="height:4px"><div class="progress-bar bg-warning" style="width:' + cpu + '%"></div></div>') +
        this._statCard('bi-lightning-charge text-primary', '活跃协程', Fmt.number(activeG), '<small class="text-muted">总计: ' + Fmt.number(totalG) + '</small>') +
        this._statCard('bi-clock text-success', '运行时间', Fmt.duration(uptime * 1000), '<small class="text-muted">启动: ' + Fmt.dateTime(d.startup_time || Date.now()) + '</small>');

      var si = (d.system_info || {});
      document.getElementById('sys-info').innerHTML =
        '<div class="card-header d-flex justify-content-between"><span>系统信息</span><i class="bi bi-info-circle"></i></div>' +
        '<div class="card-body"><dl class="row mb-0">' +
        '<dt class="col-sm-4">操作系统</dt><dd class="col-sm-8">' + esc(si.os || 'Unknown') + ' ' + esc(si.version || '') + '</dd>' +
        '<dt class="col-sm-4">系统架构</dt><dd class="col-sm-8">' + esc(si.arch || 'Unknown') + '</dd>' +
        '<dt class="col-sm-4">Go版本</dt><dd class="col-sm-8">' + esc(si.go_version || 'Unknown') + '</dd>' +
        '<dt class="col-sm-4">CPU核心数</dt><dd class="col-sm-8">' + (si.num_cpu || 0) + '</dd>' +
        '<dt class="col-sm-4">启动时间</dt><dd class="col-sm-8">' + Fmt.dateTime(d.startup_time) + '</dd>' +
        '<dt class="col-sm-4">进程 ID</dt><dd class="col-sm-8">' + (d.pid || 'N/A') + '</dd>' +
        '</dl></div>';

      document.getElementById('sys-runtime').innerHTML =
        '<div class="card-header">运行时统计</div><div class="card-body p-0">' +
        '<table class="table table-sm mb-0"><thead><tr><th>指标</th><th>当前值</th></tr></thead><tbody>' +
        '<tr><td>内存使用</td><td>' + Fmt.bytes(mem) + '</td></tr>' +
        '<tr><td>CPU使用率</td><td>' + cpu.toFixed(2) + '%</td></tr>' +
        '<tr><td>活跃协程数</td><td>' + Fmt.number(activeG) + '</td></tr>' +
        '<tr><td>运行时间</td><td>' + Fmt.duration(uptime * 1000) + '</td></tr>' +
        '<tr><td>总请求数</td><td>' + Fmt.number(totalReq) + '</td></tr>' +
        '<tr><td>成功率</td><td>' + successRate.toFixed(2) + '%</td></tr>' +
        '<tr><td>GC次数</td><td>' + Fmt.number((d.memory && d.memory.num_gc) || 0) + '</td></tr>' +
        '</tbody></table></div>';
    },
    _statCard: function (icon, label, value, extra) {
      return '<div class="col-sm-6 col-lg-3"><div class="card stat-card"><div class="card-body">' +
        '<div class="d-flex align-items-center gap-3"><i class="bi ' + icon + ' stat-icon"></i><div><div class="stat-value">' + esc(value) + '</div><div class="stat-label">' + esc(label) + '</div></div></div>' +
        (extra || '') + '</div></div></div>';
    }
  };

  // ===========================
  // Heap Page
  // ===========================
  var HeapPage = {
    interval: null, state: { search: '', minBytes: '', minObjects: '', page: 1, pageSize: 50, sortCol: '', sortDir: 'asc' }, data: null,
    render: function () {
      document.getElementById('app-content').innerHTML = '<h2 class="page-title"><i class="bi bi-database"></i>堆内存分析</h2><div id="heap-summary" class="row g-2 mb-3"></div><div id="heap-controls"></div><div class="card"><div class="card-body p-0" id="heap-table"></div></div>';
      this._renderControls();
      this._load();
      var self = this;
      this.interval = setInterval(function () { self._load(); }, 30000);
    },
    destroy: function () { if (this.interval) { clearInterval(this.interval); this.interval = null; } },
    _renderControls: function () {
      document.getElementById('heap-controls').innerHTML =
        '<div class="controls-wrapper">' +
        '<div class="input-group" style="width:280px"><span class="input-group-text"><i class="bi bi-search"></i></span><input type="text" class="form-control" id="h-search" placeholder="搜索函数名或文件路径"></div>' +
        '<input type="number" class="form-control" id="h-min-bytes" placeholder="最小内存(字节)" style="width:160px" min="0">' +
        '<input type="number" class="form-control" id="h-min-objects" placeholder="最小对象数" style="width:140px" min="0">' +
        '<button class="btn btn-primary" id="h-refresh"><i class="bi bi-arrow-clockwise me-1"></i>刷新数据</button>' +
        '<span class="text-muted small" id="h-count"></span>' +
        '</div>';
      var self = this;
      document.getElementById('h-search').addEventListener('input', function (e) { self.state.search = e.target.value; self.state.page = 1; self._renderTable(); });
      document.getElementById('h-min-bytes').addEventListener('input', function (e) { self.state.minBytes = e.target.value; self.state.page = 1; self._renderTable(); });
      document.getElementById('h-min-objects').addEventListener('input', function (e) { self.state.minObjects = e.target.value; self.state.page = 1; self._renderTable(); });
      document.getElementById('h-refresh').addEventListener('click', function () { self._load(); });
    },
    _load: function () {
      var self = this;
      Api.getHeapProfile().then(function (d) { self.data = d; self._renderSummary(); self._renderTable(); }).catch(function (e) { console.error(e); });
    },
    _renderSummary: function () {
      if (!this.data) return;
      var d = this.data;
      document.getElementById('heap-summary').innerHTML =
        '<div class="col-6 col-md-3"><div class="card"><div class="card-body text-center py-2"><div class="fs-6 fw-semibold text-primary">' + d.total_inuse_objects.toLocaleString() + '</div><small class="text-muted">使用中对象</small></div></div></div>' +
        '<div class="col-6 col-md-3"><div class="card"><div class="card-body text-center py-2"><div class="fs-6 fw-semibold text-success">' + Fmt.bytes(d.total_inuse_bytes) + '</div><small class="text-muted">使用中内存</small></div></div></div>' +
        '<div class="col-6 col-md-3"><div class="card"><div class="card-body text-center py-2"><div class="fs-6 fw-semibold text-warning">' + d.total_alloc_objects.toLocaleString() + '</div><small class="text-muted">总分配对象</small></div></div></div>' +
        '<div class="col-6 col-md-3"><div class="card"><div class="card-body text-center py-2"><div class="fs-6 fw-semibold text-danger">' + Fmt.bytes(d.total_alloc_bytes) + '</div><small class="text-muted">总分配内存</small></div></div></div>';
    },
    _getFiltered: function () {
      if (!this.data || !this.data.entries) return [];
      var s = this.state;
      return this.data.entries.filter(function (item) {
        var matchSearch = !s.search || (item.top_function || '').toLowerCase().indexOf(s.search.toLowerCase()) >= 0 ||
          (item.stack_trace || []).some(function (st) { return st.toLowerCase().indexOf(s.search.toLowerCase()) >= 0; });
        var matchBytes = !s.minBytes || item.inuse_bytes >= parseInt(s.minBytes);
        var matchObjects = !s.minObjects || item.inuse_objects >= parseInt(s.minObjects);
        return matchSearch && matchBytes && matchObjects;
      });
    },
    _renderTable: function () {
      var filtered = this._getFiltered();
      var totalEntries = this.data ? this.data.entries.length : 0;
      var s = this.state;

      var countEl = document.getElementById('h-count');
      if (countEl) countEl.textContent = '显示 ' + filtered.length + ' / ' + totalEntries + ' 条记录';

      if (s.sortCol) {
        filtered = filtered.slice().sort(function (a, b) {
          var va = a[s.sortCol] || 0, vb = b[s.sortCol] || 0;
          return s.sortDir === 'desc' ? vb - va : va - vb;
        });
      }
      var total = filtered.length;
      var start = (s.page - 1) * s.pageSize;
      var pageData = filtered.slice(start, start + s.pageSize);
      var totalInuse = this.data ? this.data.total_inuse_bytes : 1;

      var sortCols = ['inuse_objects', 'inuse_bytes', 'alloc_objects', 'alloc_bytes'];
      var h = '<div class="table-responsive"><table class="table table-sm table-striped table-hover mb-0"><thead><tr>' +
        '<th style="width:400px">函数</th>';
      [['inuse_objects', '使用中对象'], ['inuse_bytes', '使用中内存'], ['alloc_objects', '总分配对象'], ['alloc_bytes', '总分配内存']].forEach(function (c) {
        h += '<th class="text-end th-sortable' + (s.sortCol === c[0] ? ' sort-' + s.sortDir : '') + '" data-sort="' + c[0] + '" style="width:120px">' + c[1] + ' <i class="bi bi-arrow-down-up sort-icon"></i></th>';
      });
      h += '<th class="text-center" style="width:120px">内存占比</th><th class="text-center" style="width:80px">操作</th></tr></thead><tbody>';

      if (pageData.length === 0) {
        h += '<tr><td colspan="7"><div class="empty-state"><i class="bi bi-database"></i><p>暂无堆内存数据</p></div></td></tr>';
      } else {
        pageData.forEach(function (e) {
          var stacks = parseStackTrace(e.stack_trace);
          var funcHtml = '<div class="func-cell"><div class="func-name">' + (stacks.length > 0 ? esc(stacks[0].func) : esc(e.top_function)) + '</div>';
          if (stacks.length > 0) {
            funcHtml += '<div class="stack-detail">';
            stacks.forEach(function (st) {
              funcHtml += '<div><span class="text-primary">' + esc(st.func) + '</span>';
              if (st.file) funcHtml += '<br><span class="text-muted">' + esc(st.file) + (st.line ? ':' + esc(st.line) : '') + '</span>';
              funcHtml += '</div>';
            });
            funcHtml += '</div>';
          }
          funcHtml += '</div>';

          var pct = totalInuse > 0 ? (e.inuse_bytes / totalInuse * 100) : 0;
          var pctColor = pct > 10 ? 'bg-danger' : pct > 5 ? 'bg-warning' : 'bg-success';

          h += '<tr><td>' + funcHtml + '</td>' +
            '<td class="text-end">' + e.inuse_objects.toLocaleString() + '</td>' +
            '<td class="text-end">' + Fmt.bytes(e.inuse_bytes) + '</td>' +
            '<td class="text-end">' + e.alloc_objects.toLocaleString() + '</td>' +
            '<td class="text-end">' + Fmt.bytes(e.alloc_bytes) + '</td>' +
            '<td class="text-center"><div class="progress" style="height:4px" title="' + pct.toFixed(2) + '%"><div class="progress-bar ' + pctColor + '" style="width:' + pct + '%"></div></div></td>' +
            '<td class="text-center"><button class="btn btn-sm btn-link heap-view" data-idx="' + filtered.indexOf(e) + '">查看</button></td></tr>';
        });
      }
      h += '</tbody></table></div>';
      h += renderPagination(total, s.page, s.pageSize);

      document.getElementById('heap-table').innerHTML = h;
      this._bindTableEvents(filtered);
    },
    _bindTableEvents: function (filtered) {
      var self = this;
      document.querySelectorAll('#heap-table .heap-view').forEach(function (btn) {
        btn.addEventListener('click', function () { StackTraceModal.show(filtered[parseInt(btn.dataset.idx)]); });
      });
      document.querySelectorAll('#heap-table .th-sortable').forEach(function (th) {
        th.addEventListener('click', function () {
          var col = th.dataset.sort;
          if (self.state.sortCol === col) self.state.sortDir = self.state.sortDir === 'asc' ? 'desc' : 'asc';
          else { self.state.sortCol = col; self.state.sortDir = 'asc'; }
          self._renderTable();
        });
      });
      document.querySelectorAll('#heap-table .page-link[data-page]').forEach(function (a) {
        a.addEventListener('click', function (e) { e.preventDefault(); self.state.page = parseInt(a.dataset.page); self._renderTable(); });
      });
      var pss = document.querySelector('#heap-table .page-size-select');
      if (pss) pss.addEventListener('change', function (e) { self.state.pageSize = parseInt(e.target.value); self.state.page = 1; self._renderTable(); });
    }
  };

  // ===========================
  // Router
  // ===========================
  var Router = {
    pages: {
      dashboard: DashboardPage, goroutines: GoroutinesPage, requests: RequestsPage, system: SystemPage, heap: HeapPage
    },
    currentPage: null,
    init: function () {
      var self = this;
      window.addEventListener('hashchange', function () { self._route(); });
      WS.connect();
      self._route();
    },
    _route: function () {
      var hash = location.hash.replace(/^#\/?/, '') || 'dashboard';
      var pageName = hash.split('/')[0] || 'dashboard';
      if (!this.pages[pageName]) pageName = 'dashboard';

      if (this.currentPage && this.currentPage.destroy) this.currentPage.destroy();

      this.currentPage = this.pages[pageName];
      this._updateNav(pageName);
      this.currentPage.render();
    },
    _updateNav: function (pageName) {
      document.querySelectorAll('#nav-links .nav-link').forEach(function (link) {
        link.classList.toggle('active', link.dataset.page === pageName);
      });
    }
  };

  // ===========================
  // Start
  // ===========================
  function initApp() {
    var yearEl = document.getElementById('copyright-year');
    if (yearEl) yearEl.textContent = new Date().getFullYear();
    Router.init();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initApp);
  } else {
    initApp();
  }

})();
