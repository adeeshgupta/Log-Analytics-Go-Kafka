-- Sample Alert Rules Migration
-- This script creates sample alert rules for common monitoring scenarios

-- Insert sample alert rules
INSERT INTO alert_rules (name, description, `condition`, threshold, time_window, severity, enabled, created_at, updated_at) VALUES
('High Error Rate', 'Alert when error rate exceeds 5% in the last 5 minutes', 'COUNT(CASE WHEN level = ''error'' THEN 1 END) * 100.0 / COUNT(*)', 5.0, 5, 'high', true, NOW(), NOW()),
('High Response Time', 'Alert when average response time exceeds 1000ms in the last 5 minutes', 'AVG(response_time_ms)', 1000.0, 5, 'medium', true, NOW(), NOW()),
('Fatal Errors', 'Alert when any fatal errors occur in the last 5 minutes', 'COUNT(CASE WHEN level = ''fatal'' THEN 1 END)',1.0, 5, 'critical', true, NOW(), NOW()),
('High Warning Count', 'Alert when warning count exceeds 10 in the last 5 minutes', 'COUNT(CASE WHEN level = ''warn'' THEN 1 END)', 10.0, 5, 'low', true, NOW(), NOW()),
('Low Response Time', 'Alert when average response time is below 50ms (potential caching issues)', 'AVG(response_time_ms)', 50.0, 5, 'low', false, NOW(), NOW()),
('High Error Count', 'Alert when error count exceeds 5 in the last 5 minutes', 'COUNT(CASE WHEN level = ''error'' THEN 1 END)', 5.0, 5, 'high', true, NOW(), NOW()),
('Service Unavailable', 'Alert when any service has no logs in the last 10 minutes', 'COUNT(DISTINCT service)', 0.0, 10, 'critical', true, NOW(), NOW()),
('High Success Rate', 'Alert when success rate (2xx responses) is below 95% in the last 5 minutes', '(COUNT(CASE WHEN response_status >= 200 AND response_status < 300 THEN 1 END) * 100.0) / COUNT(CASE WHEN response_status IS NOT NULL THEN 1 END)', 95.0, 5, 'medium', true, NOW(), NOW());

-- Sample alert rules migration completed successfully
