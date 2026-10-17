"""Configuration regressions; these do not require a Docker engine.

Set COMPOSE_BINARY to a standalone Compose executable, or install docker compose.
"""
import json
import os
from pathlib import Path
import subprocess
import unittest

ROOT = Path(__file__).resolve().parents[1]


def compose(production=False, configured=False):
    command = [os.environ['COMPOSE_BINARY']] if os.environ.get('COMPOSE_BINARY') else ['docker', 'compose']
    command += ['--profile', '*', '-f', 'docker-compose.yml']
    if production:
        command += ['-f', 'compose.prod.yml']
    command += ['config', '--format', 'json']
    env = {k: v for k, v in os.environ.items() if not k.startswith('PRODUCTION_')}
    if configured:
        env.update({
            'PRODUCTION_DOMAIN': 'customer.example.test',
            'PRODUCTION_ADMIN_DOMAIN': 'admin.example.test',
            'PRODUCTION_DB_PASSWORD': 'test-only-database-secret',
            'PRODUCTION_JWT_SECRET': 'test-only-customer-signing-secret-32-chars',
            'PRODUCTION_ADMIN_JWT_SECRET': 'test-only-admin-signing-secret-32-chars',
            'PRODUCTION_SMTP_HOST': 'smtp.example.test',
            'PRODUCTION_SMTP_USER': 'test@example.test',
            'PRODUCTION_SMTP_PASS': 'test-only-smtp-secret',
            'PRODUCTION_SMTP_FROM': 'test@example.test',
        })
    return subprocess.run(command, cwd=ROOT, env=env, capture_output=True, text=True, timeout=30)


class ComposeConfigurationTests(unittest.TestCase):
    def test_development_internal_services_have_no_published_ports(self):
        result = compose()
        self.assertEqual(result.returncode, 0, result.stderr)
        services = json.loads(result.stdout)['services']
        for name in ['postgres', 'redis', 'user-service', 'admin-service', 'inventory-service', 'email-service', 'api-gateway']:
            self.assertFalse(services[name].get('ports'), name)
        for name in ['user-frontend', 'admin-frontend', 'mailpit']:
            self.assertTrue(all(p['host_ip'] == '127.0.0.1' for p in services[name]['ports']), name)
        self.assertEqual(services['email-service']['environment']['SMTP_MODE'], 'capture')
        self.assertEqual(services['email-service']['environment']['SMTP_HOST'], 'mailpit')

    def test_missing_production_values_fail_before_deployment(self):
        result = compose(production=True)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn('PRODUCTION_', result.stderr)

    def test_production_override_removes_frontend_ports_and_demo_bootstrap(self):
        result = compose(production=True, configured=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        services = json.loads(result.stdout)['services']
        for name in ['user-frontend', 'admin-frontend']:
            self.assertFalse(services[name].get('ports'), name)
        self.assertEqual(services['email-service']['environment']['SMTP_MODE'], 'smtp')
        self.assertEqual(services['bootstrap-admin']['environment']['APP_ENV'], 'production')
        self.assertIn('development-mail', services['mailpit']['profiles'])
        self.assertEqual({p['target'] for p in services['edge']['ports']}, {80, 443})

    def test_no_service_env_files_in_working_tree(self):
        for name in ['user-service', 'api-gateway', 'inventory-service', 'email-service']:
            self.assertFalse((ROOT / name / '.env').exists(), name)


if __name__ == '__main__':
    unittest.main()
