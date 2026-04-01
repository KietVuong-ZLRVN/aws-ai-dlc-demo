// Domain types matching the backend DTOs

export interface ComboItem {
  configSku: string
  simpleSku: string
  name: string
  imageUrl: string
  price: number
  inStock: boolean
}

export interface Combo {
  id: string
  shopperId: string
  name: string
  visibility: 'public' | 'private'
  shareToken?: string
  items: ComboItem[]
}

export interface SaveComboPayload {
  name: string
  visibility: 'public' | 'private'
  items: Array<{
    configSku: string
    simpleSku: string
    name: string
    imageUrl: string
    price: number
  }>
}
