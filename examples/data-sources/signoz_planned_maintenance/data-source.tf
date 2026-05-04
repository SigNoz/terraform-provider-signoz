data "signoz_planned_maintenance" "by_id" {
  id = signoz_planned_maintenance.weekly.id
}

output "planned_maintenance_name" {
  value = data.signoz_planned_maintenance.by_id.name
}
