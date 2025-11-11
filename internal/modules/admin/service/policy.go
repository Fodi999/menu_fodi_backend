package service

import "errors"

// AdminPolicy проверяет права доступа администратора
type AdminPolicy interface {
	// CanManageUsers проверяет право на управление пользователями
	CanManageUsers(userRole string) error

	// CanManageOrders проверяет право на управление заказами
	CanManageOrders(userRole string) error

	// CanViewStats проверяет право на просмотр статистики
	CanViewStats(userRole string) error
}

// adminPolicy реализация интерфейса AdminPolicy
type adminPolicy struct{}

// NewAdminPolicy создаёт новый экземпляр политики доступа
func NewAdminPolicy() AdminPolicy {
	return &adminPolicy{}
}

// CanManageUsers проверяет право на управление пользователями (только admin)
func (p *adminPolicy) CanManageUsers(userRole string) error {
	if userRole != "admin" {
		return errors.New("forbidden: only admins can manage users")
	}
	return nil
}

// CanManageOrders проверяет право на управление заказами (только admin)
func (p *adminPolicy) CanManageOrders(userRole string) error {
	if userRole != "admin" {
		return errors.New("forbidden: only admins can manage orders")
	}
	return nil
}

// CanViewStats проверяет право на просмотр статистики (только admin)
func (p *adminPolicy) CanViewStats(userRole string) error {
	if userRole != "admin" {
		return errors.New("forbidden: only admins can view stats")
	}
	return nil
}
