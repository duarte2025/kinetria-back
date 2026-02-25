-- Seed data for local development
-- Run with: make seed

-- Clean existing data (in reverse order of dependencies)
DELETE FROM set_records;
DELETE FROM sessions;
DELETE FROM exercises;
DELETE FROM workouts;
DELETE FROM refresh_tokens;
DELETE FROM audit_log;
DELETE FROM users;

-- Insert test users
INSERT INTO users (id, name, email, password_hash, profile_image_url, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'João Silva', 'joao@example.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYIq.Zu3aCu', 'https://i.pravatar.cc/150?img=12', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440002', 'Maria Santos', 'maria@example.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYIq.Zu3aCu', 'https://i.pravatar.cc/150?img=45', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440003', 'Pedro Costa', 'pedro@example.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYIq.Zu3aCu', 'https://i.pravatar.cc/150?img=33', NOW(), NOW());
-- Password for all users: senha123

-- Insert workouts for João
INSERT INTO workouts (id, user_id, name, description, type, intensity, created_at, updated_at) VALUES
('660e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 'Treino A - Peito e Tríceps', 'Foco em força e hipertrofia do peitoral', 'HIPERTROFIA', 'ALTA', NOW(), NOW()),
('660e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', 'Treino B - Costas e Bíceps', 'Desenvolvimento de costas e braços', 'HIPERTROFIA', 'ALTA', NOW(), NOW()),
('660e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440001', 'Treino C - Pernas', 'Treino completo de membros inferiores', 'FORCA', 'ALTA', NOW(), NOW()),
('660e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440001', 'Mobilidade Matinal', 'Alongamento e mobilidade articular', 'MOBILIDADE', 'BAIXA', NOW(), NOW());

-- Insert workouts for Maria
INSERT INTO workouts (id, user_id, name, description, type, intensity, created_at, updated_at) VALUES
('660e8400-e29b-41d4-a716-446655440005', '550e8400-e29b-41d4-a716-446655440002', 'Full Body', 'Treino de corpo inteiro 3x por semana', 'HIPERTROFIA', 'MODERADA', NOW(), NOW()),
('660e8400-e29b-41d4-a716-446655440006', '550e8400-e29b-41d4-a716-446655440002', 'HIIT Cardio', 'Treino intervalado de alta intensidade', 'CONDICIONAMENTO', 'ALTA', NOW(), NOW());

-- Insert exercises for Treino A (João)
INSERT INTO exercises (id, workout_id, name, sets, reps, weight_grams, rest_seconds, notes, order_index, created_at, updated_at) VALUES
('770e8400-e29b-41d4-a716-446655440001', '660e8400-e29b-41d4-a716-446655440001', 'Supino Reto', 4, 10, 80000, 90, 'Manter escápulas retraídas', 1, NOW(), NOW()),
('770e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440001', 'Supino Inclinado', 3, 12, 60000, 90, 'Inclinação de 30 graus', 2, NOW(), NOW()),
('770e8400-e29b-41d4-a716-446655440003', '660e8400-e29b-41d4-a716-446655440001', 'Crucifixo', 3, 15, 20000, 60, 'Foco na contração', 3, NOW(), NOW()),
('770e8400-e29b-41d4-a716-446655440004', '660e8400-e29b-41d4-a716-446655440001', 'Tríceps Testa', 3, 12, 30000, 60, 'Cotovelos fixos', 4, NOW(), NOW()),
('770e8400-e29b-41d4-a716-446655440005', '660e8400-e29b-41d4-a716-446655440001', 'Tríceps Corda', 3, 15, 40000, 60, 'Extensão completa', 5, NOW(), NOW());

-- Insert exercises for Treino B (João)
INSERT INTO exercises (id, workout_id, name, sets, reps, weight_grams, rest_seconds, notes, order_index, created_at, updated_at) VALUES
('770e8400-e29b-41d4-a716-446655440006', '660e8400-e29b-41d4-a716-446655440002', 'Barra Fixa', 4, 8, 0, 120, 'Pegada pronada', 1, NOW(), NOW()),
('770e8400-e29b-41d4-a716-446655440007', '660e8400-e29b-41d4-a716-446655440002', 'Remada Curvada', 4, 10, 70000, 90, 'Costas retas', 2, NOW(), NOW()),
('770e8400-e29b-41d4-a716-446655440008', '660e8400-e29b-41d4-a716-446655440002', 'Pulldown', 3, 12, 50000, 60, 'Puxar até o peito', 3, NOW(), NOW()),
('770e8400-e29b-41d4-a716-446655440009', '660e8400-e29b-41d4-a716-446655440002', 'Rosca Direta', 3, 12, 30000, 60, 'Sem balanço', 4, NOW(), NOW()),
('770e8400-e29b-41d4-a716-446655440010', '660e8400-e29b-41d4-a716-446655440002', 'Rosca Martelo', 3, 12, 25000, 60, 'Pegada neutra', 5, NOW(), NOW());

-- Insert exercises for Treino C (João)
INSERT INTO exercises (id, workout_id, name, sets, reps, weight_grams, rest_seconds, notes, order_index, created_at, updated_at) VALUES
('770e8400-e29b-41d4-a716-446655440011', '660e8400-e29b-41d4-a716-446655440003', 'Agachamento Livre', 5, 8, 100000, 180, 'Profundidade completa', 1, NOW(), NOW()),
('770e8400-e29b-41d4-a716-446655440012', '660e8400-e29b-41d4-a716-446655440003', 'Leg Press', 4, 12, 200000, 120, '45 graus', 2, NOW(), NOW()),
('770e8400-e29b-41d4-a716-446655440013', '660e8400-e29b-41d4-a716-446655440003', 'Cadeira Extensora', 3, 15, 50000, 60, 'Contração no topo', 3, NOW(), NOW()),
('770e8400-e29b-41d4-a716-446655440014', '660e8400-e29b-41d4-a716-446655440003', 'Mesa Flexora', 3, 15, 40000, 60, 'Controlar descida', 4, NOW(), NOW()),
('770e8400-e29b-41d4-a716-446655440015', '660e8400-e29b-41d4-a716-446655440003', 'Panturrilha em Pé', 4, 20, 80000, 60, 'Amplitude completa', 5, NOW(), NOW());

-- Insert exercises for Mobilidade (João)
INSERT INTO exercises (id, workout_id, name, sets, reps, weight_grams, rest_seconds, notes, order_index, created_at, updated_at) VALUES
('770e8400-e29b-41d4-a716-446655440016', '660e8400-e29b-41d4-a716-446655440004', 'Cat-Cow', 3, 10, 0, 30, 'Respiração profunda', 1, NOW(), NOW()),
('770e8400-e29b-41d4-a716-446655440017', '660e8400-e29b-41d4-a716-446655440004', 'World Greatest Stretch', 2, 8, 0, 30, 'Cada lado', 2, NOW(), NOW()),
('770e8400-e29b-41d4-a716-446655440018', '660e8400-e29b-41d4-a716-446655440004', 'Hip Circles', 2, 10, 0, 30, 'Ambas direções', 3, NOW(), NOW());

-- Insert exercises for Full Body (Maria)
INSERT INTO exercises (id, workout_id, name, sets, reps, weight_grams, rest_seconds, notes, order_index, created_at, updated_at) VALUES
('770e8400-e29b-41d4-a716-446655440019', '660e8400-e29b-41d4-a716-446655440005', 'Agachamento Goblet', 3, 12, 20000, 90, 'Haltere no peito', 1, NOW(), NOW()),
('770e8400-e29b-41d4-a716-446655440020', '660e8400-e29b-41d4-a716-446655440005', 'Flexão', 3, 12, 0, 60, 'Pode ser no joelho', 2, NOW(), NOW()),
('770e8400-e29b-41d4-a716-446655440021', '660e8400-e29b-41d4-a716-446655440005', 'Remada com Halteres', 3, 12, 15000, 60, 'Cada braço', 3, NOW(), NOW()),
('770e8400-e29b-41d4-a716-446655440022', '660e8400-e29b-41d4-a716-446655440005', 'Prancha', 3, 30, 0, 60, '30 segundos', 4, NOW(), NOW());

-- Insert a completed session for João (yesterday)
INSERT INTO sessions (id, user_id, workout_id, status, started_at, finished_at, notes, created_at, updated_at) VALUES
('880e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', '660e8400-e29b-41d4-a716-446655440001', 'completed', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day' + INTERVAL '1 hour', 'Treino excelente!', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day');

-- Insert set records for the completed session
INSERT INTO set_records (id, session_id, exercise_id, set_number, reps, weight_grams, status, created_at) VALUES
('990e8400-e29b-41d4-a716-446655440001', '880e8400-e29b-41d4-a716-446655440001', '770e8400-e29b-41d4-a716-446655440001', 1, 10, 80000, 'completed', NOW() - INTERVAL '1 day'),
('990e8400-e29b-41d4-a716-446655440002', '880e8400-e29b-41d4-a716-446655440001', '770e8400-e29b-41d4-a716-446655440001', 2, 10, 80000, 'completed', NOW() - INTERVAL '1 day'),
('990e8400-e29b-41d4-a716-446655440003', '880e8400-e29b-41d4-a716-446655440001', '770e8400-e29b-41d4-a716-446655440001', 3, 9, 80000, 'completed', NOW() - INTERVAL '1 day'),
('990e8400-e29b-41d4-a716-446655440004', '880e8400-e29b-41d4-a716-446655440001', '770e8400-e29b-41d4-a716-446655440001', 4, 8, 80000, 'completed', NOW() - INTERVAL '1 day'),
('990e8400-e29b-41d4-a716-446655440005', '880e8400-e29b-41d4-a716-446655440001', '770e8400-e29b-41d4-a716-446655440002', 1, 12, 60000, 'completed', NOW() - INTERVAL '1 day'),
('990e8400-e29b-41d4-a716-446655440006', '880e8400-e29b-41d4-a716-446655440001', '770e8400-e29b-41d4-a716-446655440002', 2, 12, 60000, 'completed', NOW() - INTERVAL '1 day'),
('990e8400-e29b-41d4-a716-446655440007', '880e8400-e29b-41d4-a716-446655440001', '770e8400-e29b-41d4-a716-446655440002', 3, 11, 60000, 'completed', NOW() - INTERVAL '1 day');

-- Insert another completed session for João (3 days ago)
INSERT INTO sessions (id, user_id, workout_id, status, started_at, finished_at, notes, created_at, updated_at) VALUES
('880e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', '660e8400-e29b-41d4-a716-446655440002', 'completed', NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days' + INTERVAL '55 minutes', 'Bom treino de costas', NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days');

-- Insert a completed session for Maria (2 days ago)
INSERT INTO sessions (id, user_id, workout_id, status, started_at, finished_at, notes, created_at, updated_at) VALUES
('880e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440005', 'completed', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days' + INTERVAL '45 minutes', 'Primeira semana!', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days');

SELECT 'Seed completed successfully!' as message;
SELECT 'Users created: 3' as info;
SELECT 'Workouts created: 6' as info;
SELECT 'Exercises created: 22' as info;
SELECT 'Sessions created: 3' as info;
SELECT 'Set records created: 7' as info;
SELECT '' as separator;
SELECT 'Test credentials:' as info;
SELECT '  Email: joao@example.com | maria@example.com | pedro@example.com' as credentials;
SELECT '  Password: senha123' as credentials;
