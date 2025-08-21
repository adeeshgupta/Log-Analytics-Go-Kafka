-- Sample Data Migration
-- This script creates sample log data for testing and demonstration

-- Insert sample log data for the last 24 hours
INSERT INTO logs (timestamp, level, service, message, trace_id, user_id, request_method, request_path, response_status, response_time_ms, created_at) VALUES
-- Recent logs (last hour)
(NOW() - INTERVAL 30 MINUTE, 'info', 'api-gateway', 'Request processed successfully', 'trace-001', 'user-123', 'GET', '/api/users', 200, 150, NOW() - INTERVAL 30 MINUTE),
(NOW() - INTERVAL 25 MINUTE, 'error', 'user-service', 'Database connection failed', 'trace-002', 'user-456', 'POST', '/api/users', 500, 2500, NOW() - INTERVAL 25 MINUTE),
(NOW() - INTERVAL 20 MINUTE, 'warn', 'payment-service', 'Slow response time detected', 'trace-003', 'user-789', 'PUT', '/api/payments', 200, 1200, NOW() - INTERVAL 20 MINUTE),
(NOW() - INTERVAL 15 MINUTE, 'info', 'order-service', 'Order created successfully', 'trace-004', 'user-101', 'POST', '/api/orders', 201, 300, NOW() - INTERVAL 15 MINUTE),
(NOW() - INTERVAL 10 MINUTE, 'fatal', 'notification-service', 'Critical system failure', 'trace-005', 'user-202', 'GET', '/api/notifications', 500, 3000, NOW() - INTERVAL 10 MINUTE),
(NOW() - INTERVAL 5 MINUTE, 'debug', 'api-gateway', 'Processing request', 'trace-006', 'user-303', 'DELETE', '/api/users/123', 204, 80, NOW() - INTERVAL 5 MINUTE),

-- Older logs (2-6 hours ago)
(NOW() - INTERVAL 2 HOUR, 'info', 'api-gateway', 'Request processed successfully', 'trace-007', 'user-404', 'GET', '/api/products', 200, 120, NOW() - INTERVAL 2 HOUR),
(NOW() - INTERVAL 2 HOUR + INTERVAL 5 MINUTE, 'error', 'user-service', 'Invalid request payload', 'trace-008', 'user-505', 'POST', '/api/users', 400, 200, NOW() - INTERVAL 2 HOUR + INTERVAL 5 MINUTE),
(NOW() - INTERVAL 3 HOUR, 'warn', 'payment-service', 'Payment processing delayed', 'trace-009', 'user-606', 'POST', '/api/payments', 202, 800, NOW() - INTERVAL 3 HOUR),
(NOW() - INTERVAL 4 HOUR, 'info', 'order-service', 'Order status updated', 'trace-010', 'user-707', 'PUT', '/api/orders/456', 200, 250, NOW() - INTERVAL 4 HOUR),
(NOW() - INTERVAL 5 HOUR, 'error', 'notification-service', 'Email service unavailable', 'trace-011', 'user-808', 'POST', '/api/notifications', 503, 1500, NOW() - INTERVAL 5 HOUR),
(NOW() - INTERVAL 6 HOUR, 'info', 'api-gateway', 'Health check passed', 'trace-012', 'system', 'GET', '/health', 200, 50, NOW() - INTERVAL 6 HOUR),

-- Much older logs (12-24 hours ago)
(NOW() - INTERVAL 12 HOUR, 'info', 'api-gateway', 'System startup completed', 'trace-013', 'system', 'GET', '/health', 200, 100, NOW() - INTERVAL 12 HOUR),
(NOW() - INTERVAL 18 HOUR, 'warn', 'user-service', 'High memory usage detected', 'trace-014', 'system', 'GET', '/metrics', 200, 300, NOW() - INTERVAL 18 HOUR),
(NOW() - INTERVAL 24 HOUR, 'info', 'payment-service', 'Daily reconciliation started', 'trace-015', 'system', 'POST', '/internal/reconcile', 200, 5000, NOW() - INTERVAL 24 HOUR);

-- Insert more logs to trigger alerts (high error rate scenario)
INSERT INTO logs (timestamp, level, service, message, trace_id, user_id, request_method, request_path, response_status, response_time_ms, created_at) VALUES
-- Recent errors to trigger high error rate alert
(NOW() - INTERVAL 2 MINUTE, 'error', 'api-gateway', 'Authentication failed', 'trace-016', 'user-909', 'GET', '/api/protected', 401, 200, NOW() - INTERVAL 2 MINUTE),
(NOW() - INTERVAL 1 MINUTE, 'error', 'user-service', 'User not found', 'trace-017', 'user-1010', 'GET', '/api/users/999', 404, 150, NOW() - INTERVAL 1 MINUTE),
(NOW() - INTERVAL 30 SECOND, 'error', 'payment-service', 'Payment declined', 'trace-018', 'user-1111', 'POST', '/api/payments', 400, 300, NOW() - INTERVAL 30 SECOND);

-- Insert logs with high response times to trigger latency alert
INSERT INTO logs (timestamp, level, service, message, trace_id, user_id, request_method, request_path, response_status, response_time_ms, created_at) VALUES
(NOW() - INTERVAL 45 SECOND, 'info', 'order-service', 'Complex order processing', 'trace-019', 'user-1212', 'POST', '/api/orders/complex', 200, 1500, NOW() - INTERVAL 45 SECOND),
(NOW() - INTERVAL 15 SECOND, 'info', 'notification-service', 'Bulk notification sent', 'trace-020', 'user-1313', 'POST', '/api/notifications/bulk', 200, 1800, NOW() - INTERVAL 15 SECOND);

-- Sample data migration completed successfully 
