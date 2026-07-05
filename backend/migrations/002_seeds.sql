-- ============================================================
-- 002_seeds.sql
-- 初始種子資料：後台管理員 + App 測試使用者
-- ============================================================

-- 後台管理員：admin001 / admin@1234
INSERT INTO backend_users (username, password_hash, email, role)
VALUES ('admin001', '$2a$10$9D/Tw7aaSWESQuHvN2xk6uLnUcLq.3YrAuK7tOfh.egmY9MOFSov.', 'admin001@example.com', 'admin')
ON CONFLICT (username) DO NOTHING;

-- App 測試使用者：testdemo001~003 / a12345678
INSERT INTO app_users (username, password_hash, email)
VALUES
    ('testdemo001', '$2a$10$ud.WrVNEaq2pgpNc11cWuOCujWPxOojWwrqgt2A4U/qselpu5mafS', 'testdemo001@example.com'),
    ('testdemo002', '$2a$10$ud.WrVNEaq2pgpNc11cWuOCujWPxOojWwrqgt2A4U/qselpu5mafS', 'testdemo002@example.com'),
    ('testdemo003', '$2a$10$ud.WrVNEaq2pgpNc11cWuOCujWPxOojWwrqgt2A4U/qselpu5mafS', 'testdemo003@example.com')
ON CONFLICT (username) DO NOTHING;
