resource "autonomi_attachment" "attachment"{
  workspace_id = autonomi_workspace.created_workspace.id
  node_id = autonomi_cloud_node.created_cloud_node.id
  transport_id = autonomi_transport.created_transport.id
}