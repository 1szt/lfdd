# apiMenu 模块

菜单接口：提供菜品分类与商品数据（当前为内置模拟数据），挂载于 `/api/menu`。须在 `web.Start()` 之后调用。

# 接口列表

GET /api/menu 获取菜单

# 接口说明

## GET /api/menu

返回：

```json
{
  "code": 200,
  "message": "success",
  "categories": [
    {
      "id": 1,
      "name": "热销推荐",
      "goods": [
        { "id": 101, "name": "招牌炒饭", "icon": "🍚", "desc": "大火爆炒 粒粒分明", "price": 18 }
      ]
    }
  ]
}
```

# 数据结构

CategoryItem - 分类：ID、Name、Goods（商品列表）
GoodsItem - 商品：ID、Name、Icon、Desc、Price（单位：元）

# 备注

数据为模块内硬编码的模拟数据 `menuData`，共 4 个分类（热销推荐/经典套餐/凉菜小食/饮品甜点），后续接入数据库时替换 GetMenu 的数据来源即可。
