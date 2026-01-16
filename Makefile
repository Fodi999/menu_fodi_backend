.PHONY: help login-user login-admin test-egg test-notifications build deploy

# Default target
help:
	@echo "=========================================="
	@echo "🚀 Backend Development Commands"
	@echo "=========================================="
	@echo ""
	@echo "Authentication:"
	@echo "  make login-user        - Login as home_chef (fodi85@gmail.ru)"
	@echo "  make login-admin       - Login as super_admin"
	@echo ""
	@echo "E2E Tests:"
	@echo "  make test-egg          - Test full 'Яичница' scenario"
	@echo "  make test-notifications - Test notification lifecycle"
	@echo ""
	@echo "Build & Deploy:"
	@echo "  make build             - Build server binary"
	@echo "  make deploy            - Commit and push to production"
	@echo ""
	@echo "Database:"
	@echo "  make db-url            - Show database connection string"
	@echo ""

# ============================================
# AUTHENTICATION
# ============================================

login-user:
	@chmod +x scripts/login_user.sh
	@./scripts/login_user.sh

login-admin:
	@chmod +x scripts/login_admin.sh
	@./scripts/login_admin.sh

# ============================================
# E2E TESTS
# ============================================

test-egg:
	@chmod +x scripts/test_egg_scenario.sh
	@./scripts/test_egg_scenario.sh

test-notifications:
	@chmod +x test_delete_discard_notifications.sh
	@./test_delete_discard_notifications.sh

# ============================================
# BUILD & DEPLOY
# ============================================

build:
	@echo "🔨 Building server..."
	@go build -o bin/server ./cmd/server
	@echo "✅ Build complete: bin/server"

deploy:
	@echo "🚀 Deploying to production..."
	@git add -A
	@git status --short
	@read -p "Commit message: " msg; \
	git commit -m "$$msg"
	@git push origin main
	@echo "✅ Deployed! Koyeb will auto-deploy in ~90 seconds"

# ============================================
# DATABASE
# ============================================

db-url:
	@echo "📊 Database URL:"
	@echo "$(DATABASE_URL)"

# ============================================
# CLEANUP
# ============================================

clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -f bin/server
	@rm -f bin/server_test
	@echo "✅ Clean complete"
