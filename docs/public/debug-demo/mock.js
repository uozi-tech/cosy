(function () {
  'use strict';

  var now = Math.floor(Date.now() / 1000);
  var startupTime = now - 3600;

  // ===========================
  // Mock Data
  // ===========================

  var goroutines = [
    {
      id: 'kernel-1', name: 'user-notification-sender', status: 'running',
      start_time: now - 120, end_time: 0,
      stack: 'goroutine 42 [running]:\nmain.sendNotification()\n\t/app/services/notification.go:85\nkernel.Run.func1()\n\t/app/kernel/run.go:32',
      session_logs: JSON.stringify([
        { time: now - 120, level: 0, message: 'Starting notification batch', caller: 'notification.go:85' },
        { time: now - 115, level: 0, message: '[1.253ms] [rows:300] SELECT * FROM "notifications" WHERE "notifications"."status" = \'pending\' AND "notifications"."deleted_at" IS NULL ORDER BY "notifications"."priority" DESC', caller: 'notification.go:92' },
        { time: now - 100, level: 0, message: '[0.412ms] [rows:1] UPDATE "notifications" SET "status"=\'sending\',"updated_at"=\'2026-03-05 10:30:00\' WHERE "id" = 1024', caller: 'notification.go:105' },
        { time: now - 60, level: 0, message: 'Processed 150/300 notifications', caller: 'notification.go:112' },
        { time: now - 55, level: 0, message: '[0.387ms] [rows:150] UPDATE "notifications" SET "status"=\'sent\',"sent_at"=\'2026-03-05 10:31:05\' WHERE "id" IN (1024,1025,1026,1027,...)', caller: 'notification.go:118' }
      ])
    },
    {
      id: 'kernel-2', name: 'cache-warmer', status: 'running',
      start_time: now - 300, end_time: 0,
      stack: 'goroutine 58 [running]:\nmain.warmCache()\n\t/app/services/cache.go:42\nkernel.Run.func1()\n\t/app/kernel/run.go:32',
      session_logs: JSON.stringify([
        { time: now - 300, level: 0, message: 'Cache warming started', caller: 'cache.go:42' },
        { time: now - 295, level: 0, message: '[3.782ms] [rows:512] SELECT * FROM "products" WHERE "products"."is_active" = true AND "products"."deleted_at" IS NULL LIMIT 512', caller: 'cache.go:55' },
        { time: now - 250, level: 0, message: '[2.156ms] [rows:256] SELECT * FROM "categories" WHERE "categories"."deleted_at" IS NULL', caller: 'cache.go:62' },
        { time: now - 210, level: 1, message: '[215.432ms] [rows:256] SLOW SQL >= 200ms SELECT "products"."id","products"."name","products"."price","categories"."name" FROM "products" LEFT JOIN "categories" ON "categories"."id" = "products"."category_id" WHERE "products"."deleted_at" IS NULL', caller: 'cache.go:70' },
        { time: now - 200, level: 0, message: 'Loaded 1024 entries into cache', caller: 'cache.go:78' }
      ])
    },
    {
      id: 'kernel-3', name: 'db-health-checker', status: 'running',
      start_time: now - 600, end_time: 0,
      stack: 'goroutine 63 [running]:\nmain.checkDBHealth()\n\t/app/services/health.go:28\nkernel.Run.func1()\n\t/app/kernel/run.go:32',
      session_logs: JSON.stringify([
        { time: now - 600, level: 0, message: 'Running database health check', caller: 'health.go:28' },
        { time: now - 599, level: 0, message: '[0.892ms] [rows:1] SELECT 1', caller: 'health.go:35' },
        { time: now - 300, level: 0, message: '[0.751ms] [rows:1] SELECT 1', caller: 'health.go:35' },
        { time: now - 10, level: 0, message: '[0.823ms] [rows:1] SELECT 1', caller: 'health.go:35' }
      ])
    },
    {
      id: 'kernel-4', name: 'metrics-reporter', status: 'completed',
      start_time: now - 900, end_time: now - 850,
      stack: 'goroutine 71 [finished]:\nmain.reportMetrics()\n\t/app/services/metrics.go:55\nkernel.Run.func1()\n\t/app/kernel/run.go:32',
      session_logs: JSON.stringify([
        { time: now - 900, level: 0, message: 'Collecting metrics', caller: 'metrics.go:55' },
        { time: now - 860, level: 0, message: 'Reported 42 metrics to server', caller: 'metrics.go:88' },
        { time: now - 850, level: 0, message: 'Metrics report completed', caller: 'metrics.go:95' }
      ])
    },
    {
      id: 'kernel-5', name: 'email-queue-processor', status: 'completed',
      start_time: now - 1200, end_time: now - 1100,
      stack: 'goroutine 82 [finished]:\nmain.processEmailQueue()\n\t/app/services/email.go:34\nkernel.Run.func1()\n\t/app/kernel/run.go:32',
      session_logs: JSON.stringify([
        { time: now - 1200, level: 0, message: 'Processing email queue', caller: 'email.go:34' },
        { time: now - 1195, level: 0, message: '[0.934ms] [rows:8] SELECT * FROM "email_queue" WHERE "email_queue"."status" = \'pending\' AND "email_queue"."deleted_at" IS NULL ORDER BY "email_queue"."created_at" ASC LIMIT 50', caller: 'email.go:41' },
        { time: now - 1160, level: 0, message: '[0.215ms] [rows:1] UPDATE "email_queue" SET "status"=\'sent\',"sent_at"=\'2026-03-05 09:50:40\' WHERE "id" = 301', caller: 'email.go:58' },
        { time: now - 1150, level: 0, message: 'Sent 8 emails successfully', caller: 'email.go:67' },
        { time: now - 1110, level: 0, message: '[0.178ms] [rows:8] UPDATE "email_queue" SET "status"=\'sent\' WHERE "id" IN (301,302,303,304,305,306,307,308)', caller: 'email.go:69' },
        { time: now - 1100, level: 0, message: 'Queue processing complete', caller: 'email.go:72' }
      ])
    },
    {
      id: 'kernel-6', name: 'data-sync-worker', status: 'failed',
      start_time: now - 500, end_time: now - 480,
      error: 'connection refused: tcp 10.0.1.5:5432',
      stack: 'goroutine 91 [finished]:\nmain.syncData()\n\t/app/services/sync.go:48\nkernel.Run.func1()\n\t/app/kernel/run.go:32',
      session_logs: JSON.stringify([
        { time: now - 500, level: 0, message: 'Starting data sync', caller: 'sync.go:48' },
        { time: now - 498, level: 2, message: '[3012.451ms] [rows:0] dial tcp 10.0.1.5:5432: connect: connection refused', caller: 'sync.go:55' },
        { time: now - 490, level: 1, message: 'Retrying connection (attempt 2/3)', caller: 'sync.go:62' },
        { time: now - 487, level: 2, message: '[3008.772ms] [rows:0] dial tcp 10.0.1.5:5432: connect: connection refused', caller: 'sync.go:55' },
        { time: now - 483, level: 1, message: 'Retrying connection (attempt 3/3)', caller: 'sync.go:62' },
        { time: now - 480, level: 2, message: 'Failed to connect to database: connection refused', caller: 'sync.go:70' }
      ])
    },
    {
      id: 'kernel-7', name: 'log-rotator', status: 'completed',
      start_time: now - 1800, end_time: now - 1780,
      stack: 'goroutine 35 [finished]:\nmain.rotateLogs()\n\t/app/services/logrotate.go:22\nkernel.Run.func1()\n\t/app/kernel/run.go:32',
      session_logs: JSON.stringify([
        { time: now - 1800, level: 0, message: 'Rotating log files', caller: 'logrotate.go:22' },
        { time: now - 1780, level: 0, message: 'Rotated 3 log files, freed 128MB', caller: 'logrotate.go:45' }
      ])
    },
    {
      id: 'kernel-8', name: 'session-cleanup', status: 'running',
      start_time: now - 60, end_time: 0,
      stack: 'goroutine 105 [running]:\nmain.cleanupSessions()\n\t/app/services/session.go:19\nkernel.Run.func1()\n\t/app/kernel/run.go:32',
      session_logs: JSON.stringify([
        { time: now - 60, level: 0, message: 'Cleaning expired sessions', caller: 'session.go:19' },
        { time: now - 58, level: 0, message: '[1.456ms] [rows:37] DELETE FROM "sessions" WHERE "sessions"."expires_at" < \'2026-03-05 10:29:02\' AND "sessions"."deleted_at" IS NULL', caller: 'session.go:28' },
        { time: now - 55, level: 0, message: 'Deleted 37 expired sessions', caller: 'session.go:32' }
      ])
    }
  ];

  var requests = [
    {
      request_id: 'req-a1b2c3', req_method: 'GET', req_url: '/api/users',
      resp_status_code: 200, ip: '192.168.1.100', latency: '2.35ms',
      start_time: now - 5, end_time: now - 5, user_agent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)',
      req_header: '{"Accept":"application/json","Authorization":"Bearer eyJ..."}',
      resp_header: '{"Content-Type":"application/json","X-Request-Id":"req-a1b2c3"}',
      resp_body: '{"data":[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}],"total":2}',
      session_logs: JSON.stringify([
        { time: now - 5, level: 0, message: 'GET /api/users', caller: 'handler.go:42' },
        { time: now - 5, level: 0, message: '[0.523ms] [rows:2] SELECT * FROM "users" WHERE "users"."deleted_at" IS NULL ORDER BY "users"."id" ASC', caller: 'handler.go:50' },
        { time: now - 5, level: 0, message: '[0.187ms] [rows:1] SELECT count(*) FROM "users" WHERE "users"."deleted_at" IS NULL', caller: 'handler.go:52' },
        { time: now - 5, level: 0, message: 'Query returned 2 users', caller: 'handler.go:58' }
      ])
    },
    {
      request_id: 'req-d4e5f6', req_method: 'POST', req_url: '/api/users',
      resp_status_code: 201, ip: '192.168.1.101', latency: '15.82ms',
      start_time: now - 12, end_time: now - 12, user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
      req_header: '{"Content-Type":"application/json","Authorization":"Bearer eyJ..."}',
      req_body: '{"name":"Charlie","email":"charlie@example.com","role":"editor"}',
      resp_body: '{"id":3,"name":"Charlie","email":"charlie@example.com"}',
      session_logs: JSON.stringify([
        { time: now - 12, level: 0, message: 'POST /api/users', caller: 'handler.go:85' },
        { time: now - 12, level: 0, message: '[0.341ms] [rows:0] SELECT * FROM "users" WHERE "users"."email" = \'charlie@example.com\' AND "users"."deleted_at" IS NULL LIMIT 1', caller: 'handler.go:90' },
        { time: now - 12, level: 0, message: '[1.872ms] [rows:1] INSERT INTO "users" ("name","email","role","created_at","updated_at") VALUES (\'Charlie\',\'charlie@example.com\',\'editor\',\'2026-03-05 10:29:48\',\'2026-03-05 10:29:48\')', caller: 'handler.go:97' },
        { time: now - 12, level: 0, message: 'Created user id=3', caller: 'handler.go:102' }
      ])
    },
    {
      request_id: 'req-g7h8i9', req_method: 'GET', req_url: '/api/products?page=1&limit=20',
      resp_status_code: 200, ip: '10.0.0.55', latency: '8.41ms',
      start_time: now - 20, end_time: now - 20, user_agent: 'PostmanRuntime/7.32.3',
      session_logs: JSON.stringify([
        { time: now - 20, level: 0, message: 'GET /api/products?page=1&limit=20', caller: 'product_handler.go:30' },
        { time: now - 20, level: 0, message: '[2.134ms] [rows:20] SELECT * FROM "products" WHERE "products"."is_active" = true AND "products"."deleted_at" IS NULL ORDER BY "products"."created_at" DESC LIMIT 20', caller: 'product_handler.go:45' },
        { time: now - 20, level: 0, message: '[0.892ms] [rows:1] SELECT count(*) FROM "products" WHERE "products"."is_active" = true AND "products"."deleted_at" IS NULL', caller: 'product_handler.go:47' },
        { time: now - 20, level: 0, message: '[3.215ms] [rows:20] SELECT "categories"."id","categories"."name" FROM "categories" WHERE "categories"."id" IN (1,2,3,5,8) AND "categories"."deleted_at" IS NULL', caller: 'product_handler.go:50' }
      ])
    },
    {
      request_id: 'req-j1k2l3', req_method: 'PUT', req_url: '/api/users/42',
      resp_status_code: 200, ip: '192.168.1.100', latency: '12.07ms',
      start_time: now - 35, end_time: now - 35, user_agent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)',
      req_body: '{"name":"Alice Updated","role":"admin"}',
      session_logs: JSON.stringify([
        { time: now - 35, level: 0, message: 'PUT /api/users/42', caller: 'handler.go:120' },
        { time: now - 35, level: 0, message: '[0.456ms] [rows:1] SELECT * FROM "users" WHERE "users"."id" = 42 AND "users"."deleted_at" IS NULL LIMIT 1', caller: 'handler.go:125' },
        { time: now - 35, level: 0, message: '[1.234ms] [rows:1] UPDATE "users" SET "name"=\'Alice Updated\',"role"=\'admin\',"updated_at"=\'2026-03-05 10:29:25\' WHERE "users"."id" = 42', caller: 'handler.go:133' },
        { time: now - 35, level: 0, message: 'Updated user id=42', caller: 'handler.go:138' }
      ])
    },
    {
      request_id: 'req-m4n5o6', req_method: 'DELETE', req_url: '/api/sessions/expired',
      resp_status_code: 204, ip: '10.0.0.1', latency: '3.21ms',
      start_time: now - 50, end_time: now - 50, user_agent: 'curl/8.1.2',
      session_logs: JSON.stringify([
        { time: now - 50, level: 0, message: 'DELETE /api/sessions/expired', caller: 'session_handler.go:65' },
        { time: now - 50, level: 0, message: '[2.087ms] [rows:12] DELETE FROM "sessions" WHERE "sessions"."expires_at" < \'2026-03-05 10:29:10\'', caller: 'session_handler.go:72' }
      ])
    },
    {
      request_id: 'req-p7q8r9', req_method: 'GET', req_url: '/api/orders/99999',
      resp_status_code: 404, ip: '192.168.1.200', latency: '1.05ms',
      start_time: now - 65, end_time: now - 65, user_agent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0)',
      resp_body: '{"error":"order not found"}',
      session_logs: JSON.stringify([
        { time: now - 65, level: 0, message: 'GET /api/orders/99999', caller: 'order_handler.go:38' },
        { time: now - 65, level: 2, message: '[0.312ms] [rows:0] record not found SELECT * FROM "orders" WHERE "orders"."id" = 99999 AND "orders"."deleted_at" IS NULL LIMIT 1', caller: 'order_handler.go:45' },
        { time: now - 65, level: 1, message: 'Order 99999 not found', caller: 'order_handler.go:50' }
      ])
    },
    {
      request_id: 'req-s1t2u3', req_method: 'POST', req_url: '/api/reports/generate',
      resp_status_code: 500, ip: '192.168.1.150', latency: '523.67ms',
      start_time: now - 80, end_time: now - 79, user_agent: 'Mozilla/5.0 (X11; Linux x86_64)',
      req_body: '{"type":"monthly","month":"2026-02"}',
      resp_body: '{"error":"internal server error"}',
      error: 'panic: runtime error: index out of range [5] with length 3',
      call_stack: 'main.generateReport()\n\t/app/services/report.go:142\nmain.handleReportGenerate()\n\t/app/handlers/report.go:28',
      session_logs: JSON.stringify([
        { time: now - 80, level: 0, message: 'Generating monthly report for 2026-02', caller: 'report.go:28' },
        { time: now - 80, level: 0, message: '[45.678ms] [rows:1523] SELECT * FROM "orders" WHERE "orders"."created_at" BETWEEN \'2026-02-01\' AND \'2026-02-28\' AND "orders"."deleted_at" IS NULL', caller: 'report.go:55' },
        { time: now - 80, level: 0, message: '[12.345ms] [rows:89] SELECT "products"."category_id", SUM("order_items"."quantity") as total, SUM("order_items"."price" * "order_items"."quantity") as revenue FROM "order_items" JOIN "products" ON "products"."id" = "order_items"."product_id" GROUP BY "products"."category_id"', caller: 'report.go:78' },
        { time: now - 79, level: 1, message: '[356.912ms] [rows:15000] SLOW SQL >= 200ms SELECT "orders"."id","orders"."user_id","users"."name","orders"."total" FROM "orders" LEFT JOIN "users" ON "users"."id" = "orders"."user_id" WHERE "orders"."created_at" >= \'2026-01-01\'', caller: 'report.go:110' },
        { time: now - 79, level: 2, message: 'Report generation failed: index out of range', caller: 'report.go:142' }
      ])
    },
    {
      request_id: 'req-v4w5x6', req_method: 'GET', req_url: '/api/dashboard/stats',
      resp_status_code: 200, ip: '192.168.1.100', latency: '45.23ms',
      start_time: now - 100, end_time: now - 100, user_agent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)',
      session_logs: JSON.stringify([
        { time: now - 100, level: 0, message: 'GET /api/dashboard/stats', caller: 'dashboard_handler.go:18' },
        { time: now - 100, level: 0, message: '[0.456ms] [rows:1] SELECT count(*) FROM "users" WHERE "users"."deleted_at" IS NULL', caller: 'dashboard_handler.go:25' },
        { time: now - 100, level: 0, message: '[1.234ms] [rows:1] SELECT count(*) FROM "orders" WHERE "orders"."created_at" >= \'2026-03-01\' AND "orders"."deleted_at" IS NULL', caller: 'dashboard_handler.go:28' },
        { time: now - 100, level: 0, message: '[8.567ms] [rows:1] SELECT COALESCE(SUM("total"), 0) as revenue FROM "orders" WHERE "orders"."status" = \'completed\' AND "orders"."created_at" >= \'2026-03-01\' AND "orders"."deleted_at" IS NULL', caller: 'dashboard_handler.go:31' },
        { time: now - 100, level: 0, message: '[25.891ms] [rows:30] SELECT DATE("created_at") as date, count(*) as count FROM "orders" WHERE "orders"."created_at" >= \'2026-02-03\' GROUP BY DATE("created_at") ORDER BY date', caller: 'dashboard_handler.go:38' }
      ])
    },
    {
      request_id: 'req-y7z8a1', req_method: 'PATCH', req_url: '/api/settings/notifications',
      resp_status_code: 200, ip: '192.168.1.101', latency: '6.78ms',
      start_time: now - 130, end_time: now - 130, user_agent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
      req_body: '{"email_enabled":true,"push_enabled":false}',
      session_logs: JSON.stringify([
        { time: now - 130, level: 0, message: 'PATCH /api/settings/notifications', caller: 'settings_handler.go:42' },
        { time: now - 130, level: 0, message: '[0.312ms] [rows:1] SELECT * FROM "user_settings" WHERE "user_settings"."user_id" = 1 AND "user_settings"."deleted_at" IS NULL LIMIT 1', caller: 'settings_handler.go:48' },
        { time: now - 130, level: 0, message: '[0.567ms] [rows:1] UPDATE "user_settings" SET "email_enabled"=true,"push_enabled"=false,"updated_at"=\'2026-03-05 10:27:50\' WHERE "user_settings"."user_id" = 1', caller: 'settings_handler.go:55' }
      ])
    },
    {
      request_id: 'req-b2c3d4', req_method: 'POST', req_url: '/api/auth/login',
      resp_status_code: 200, ip: '203.0.113.50', latency: '89.12ms',
      start_time: now - 160, end_time: now - 160, user_agent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)',
      req_body: '{"username":"admin","password":"***"}',
      resp_body: '{"token":"eyJhbGciOiJIUzI1NiIs...","expires_in":3600}',
      session_logs: JSON.stringify([
        { time: now - 160, level: 0, message: 'Login attempt for user: admin', caller: 'auth.go:45' },
        { time: now - 160, level: 0, message: '[0.678ms] [rows:1] SELECT * FROM "users" WHERE "users"."username" = \'admin\' AND "users"."deleted_at" IS NULL LIMIT 1', caller: 'auth.go:52' },
        { time: now - 160, level: 0, message: '[72.345ms] [rows:0] bcrypt.CompareHashAndPassword', caller: 'auth.go:60' },
        { time: now - 160, level: 0, message: '[1.123ms] [rows:1] INSERT INTO "login_logs" ("user_id","ip","user_agent","created_at") VALUES (1,\'203.0.113.50\',\'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)\',\'2026-03-05 10:27:20\')', caller: 'auth.go:68' },
        { time: now - 160, level: 0, message: 'Login successful, token issued', caller: 'auth.go:72' }
      ])
    }
  ];

  var heapEntries = [
    {
      top_function: 'runtime.malg',
      stack_trace: ['runtime.malg()\n/usr/local/go/src/runtime/proc.go:4536', 'runtime.newproc.func1()\n/usr/local/go/src/runtime/proc.go:4812', 'runtime.systemstack()\n/usr/local/go/src/runtime/asm_amd64.s:509'],
      inuse_objects: 128, inuse_bytes: 917504, alloc_objects: 256, alloc_bytes: 1835008
    },
    {
      top_function: 'bufio.NewReaderSize',
      stack_trace: ['bufio.NewReaderSize()\n/usr/local/go/src/bufio/bufio.go:60', 'net/http.newBufioReader()\n/usr/local/go/src/net/http/transport.go:1657', 'net/http.(*persistConn).readLoop()\n/usr/local/go/src/net/http/transport.go:2205'],
      inuse_objects: 64, inuse_bytes: 524288, alloc_objects: 512, alloc_bytes: 4194304
    },
    {
      top_function: 'github.com/gin-gonic/gin.(*Engine).ServeHTTP',
      stack_trace: ['github.com/gin-gonic/gin.(*Engine).ServeHTTP()\n/go/pkg/mod/github.com/gin-gonic/gin@v1.9.1/gin.go:572', 'net/http.serverHandler.ServeHTTP()\n/usr/local/go/src/net/http/server.go:2938'],
      inuse_objects: 256, inuse_bytes: 262144, alloc_objects: 8192, alloc_bytes: 8388608
    },
    {
      top_function: 'gorm.io/gorm.(*DB).Find',
      stack_trace: ['gorm.io/gorm.(*DB).Find()\n/go/pkg/mod/gorm.io/gorm@v1.25.7/finisher_api.go:165', 'main.(*UserHandler).List()\n/app/handlers/user.go:42'],
      inuse_objects: 512, inuse_bytes: 786432, alloc_objects: 4096, alloc_bytes: 6291456
    },
    {
      top_function: 'encoding/json.(*Decoder).Decode',
      stack_trace: ['encoding/json.(*Decoder).Decode()\n/usr/local/go/src/encoding/json/stream.go:55', 'github.com/gin-gonic/gin.(*Context).ShouldBindJSON()\n/go/pkg/mod/github.com/gin-gonic/gin@v1.9.1/context.go:773'],
      inuse_objects: 32, inuse_bytes: 65536, alloc_objects: 1024, alloc_bytes: 2097152
    },
    {
      top_function: 'crypto/tls.(*Conn).readHandshake',
      stack_trace: ['crypto/tls.(*Conn).readHandshake()\n/usr/local/go/src/crypto/tls/conn.go:733', 'crypto/tls.(*Conn).HandshakeContext()\n/usr/local/go/src/crypto/tls/handshake_client.go:160'],
      inuse_objects: 16, inuse_bytes: 131072, alloc_objects: 128, alloc_bytes: 1048576
    },
    {
      top_function: 'runtime.allocm',
      stack_trace: ['runtime.allocm()\n/usr/local/go/src/runtime/proc.go:1929', 'runtime.newm()\n/usr/local/go/src/runtime/proc.go:2366'],
      inuse_objects: 8, inuse_bytes: 65536, alloc_objects: 32, alloc_bytes: 262144
    }
  ];

  var totalInuseBytes = heapEntries.reduce(function (s, e) { return s + e.inuse_bytes; }, 0);
  var totalInuseObjects = heapEntries.reduce(function (s, e) { return s + e.inuse_objects; }, 0);
  var totalAllocBytes = heapEntries.reduce(function (s, e) { return s + e.alloc_bytes; }, 0);
  var totalAllocObjects = heapEntries.reduce(function (s, e) { return s + e.alloc_objects; }, 0);

  var systemInfo = {
    system_info: {
      os: 'linux', version: 'Ubuntu 22.04.3 LTS', arch: 'amd64',
      go_version: 'go1.23.4', num_cpu: 8
    },
    memory: {
      alloc: 12582912, total_alloc: 67108864, sys: 33554432,
      heap_alloc: 12582912, heap_sys: 25165824, heap_idle: 8388608,
      heap_inuse: 16777216, heap_released: 4194304, heap_objects: 48576,
      num_gc: 142
    },
    goroutines: { total: 48 },
    startup_time: startupTime,
    timestamp: now,
    pid: 12847
  };

  // ===========================
  // URL matcher
  // ===========================
  function matchEndpoint(url) {
    var u = url.split('?')[0];
    if (u.endsWith('/system')) return 'system';
    if (u.endsWith('/stats')) return 'stats';
    if (u.endsWith('/heap')) return 'heap';
    if (u.match(/\/goroutines$/)) return 'goroutines';
    if (u.match(/\/goroutine\/([^/]+)$/)) return 'goroutine_detail';
    if (u.match(/\/requests$/)) return 'requests';
    if (u.match(/\/request\/([^/]+)$/)) return 'request_detail';
    if (u.match(/\/requests\/search$/)) return 'requests';
    if (u.match(/\/connections$/)) return 'connections';
    if (u.match(/\/monitor$/)) return 'monitor';
    return null;
  }

  function extractId(url) {
    var m = url.match(/\/(goroutine|request)\/([^/?]+)/);
    return m ? decodeURIComponent(m[2]) : null;
  }

  // ===========================
  // Override fetch
  // ===========================
  var _origFetch = window.fetch;
  window.fetch = function (url, opts) {
    var endpoint = matchEndpoint(String(url));
    if (!endpoint) return _origFetch.apply(this, arguments);

    var body;
    switch (endpoint) {
      case 'system':
        body = systemInfo;
        break;
      case 'stats':
        body = {
          goroutine_stats: { active_count: 4, total_count: 48 },
          request_stats: { total_requests: requests.length, success_rate: 80.0 },
          system_stats: { memory_usage: systemInfo.memory.alloc, cpu_usage: 12.5, uptime: now - startupTime }
        };
        break;
      case 'goroutines':
        body = { data: goroutines, total: goroutines.length };
        break;
      case 'goroutine_detail':
        var gid = extractId(String(url));
        var g = goroutines.find(function (x) { return x.id === gid; });
        body = g || goroutines[0];
        break;
      case 'requests':
        body = { data: requests, total: requests.length };
        break;
      case 'request_detail':
        var rid = extractId(String(url));
        var r = requests.find(function (x) { return x.request_id === rid; });
        body = r || requests[0];
        break;
      case 'heap':
        body = {
          entries: heapEntries,
          total_inuse_objects: totalInuseObjects,
          total_inuse_bytes: totalInuseBytes,
          total_alloc_objects: totalAllocObjects,
          total_alloc_bytes: totalAllocBytes
        };
        break;
      case 'connections':
        body = { connections: 1 };
        break;
      case 'monitor':
        body = { enabled: true };
        break;
      default:
        body = {};
    }

    return Promise.resolve({
      ok: true, status: 200, statusText: 'OK',
      json: function () { return Promise.resolve(body); },
      text: function () { return Promise.resolve(JSON.stringify(body)); }
    });
  };

  // ===========================
  // Override WebSocket
  // ===========================
  window.WebSocket = function MockWebSocket() {
    var self = this;
    self.readyState = 0; // CONNECTING
    self.onopen = null;
    self.onmessage = null;
    self.onerror = null;
    self.onclose = null;
    self._timers = [];

    setTimeout(function () {
      self.readyState = 1; // OPEN
      if (self.onopen) self.onopen({ type: 'open' });

      var pushMsg = function (data) {
        if (self.readyState !== 1) return;
        if (self.onmessage) self.onmessage({ data: JSON.stringify(data) });
      };

      pushMsg({
        type: 'stats',
        data: {
          goroutine_stats: { active_count: 4, total_count: 48, completed_count: 3, failed_count: 1 },
          request_stats: { total_requests: requests.length, active_requests: 0, success_rate: 80.0 },
          system_stats: { memory_usage: systemInfo.memory.alloc, cpu_usage: 12.5, uptime: now - startupTime },
          ws_connections: 1
        }
      });

      self._timers.push(setInterval(function () {
        var n = Math.floor(Date.now() / 1000);
        var cpuJitter = 8 + Math.random() * 15;
        var memJitter = systemInfo.memory.alloc + Math.floor((Math.random() - 0.5) * 2000000);
        pushMsg({
          type: 'stats_update',
          data: {
            goroutine_stats: { active_count: 3 + Math.floor(Math.random() * 3), total_count: 48 },
            request_stats: { total_requests: requests.length, success_rate: 75 + Math.random() * 20 },
            system_stats: { memory_usage: memJitter, cpu_usage: cpuJitter, uptime: n - startupTime },
            ws_connections: 1
          }
        });
      }, 5000));
    }, 100);

    self.send = function () {};
    self.close = function () {
      self.readyState = 3;
      self._timers.forEach(clearInterval);
      if (self.onclose) self.onclose({ code: 1000, reason: 'mock close' });
    };
  };
  window.WebSocket.CONNECTING = 0;
  window.WebSocket.OPEN = 1;
  window.WebSocket.CLOSING = 2;
  window.WebSocket.CLOSED = 3;

})();
