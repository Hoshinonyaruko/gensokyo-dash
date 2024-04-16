<p align="center">
  <a href="https://www.github.com/hoshinonyaruko/gensokyo-dashboard">
    <img src="images/head.gif" width="200" height="200" alt="gensokyo">
  </a>
</p>

<div align="center">

# gensokyo-dashboard

_✨ 基于 [OneBot](https://github.com/howmanybots/onebot/blob/master/README.md) Onebotv11+QQ开发平台 机器人Dau和状态监测面板✨_  


</div>

<p align="center">
  <a href="https://raw.githubusercontent.com/hoshinonyaruko/gensokyo-dashboard/main/LICENSE">
    <img src="https://img.shields.io/github/license/hoshinonyaruko/gensokyo-dashboard" alt="license">
  </a>
  <a href="https://github.com/hoshinonyaruko/gensokyo-dashboard/releases">
    <img src="https://img.shields.io/github/v/release/hoshinonyaruko/gensokyo?color=blueviolet&include_prereleases" alt="release">
  </a>
  <a href="https://github.com/howmanybots/onebot/blob/master/README.md">
    <img src="https://img.shields.io/badge/OneBot-v11-blue?style=flat&logo=data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAEAAAABABAMAAABYR2ztAAAAIVBMVEUAAAAAAAADAwMHBwceHh4UFBQNDQ0ZGRkoKCgvLy8iIiLWSdWYAAAAAXRSTlMAQObYZgAAAQVJREFUSMftlM0RgjAQhV+0ATYK6i1Xb+iMd0qgBEqgBEuwBOxU2QDKsjvojQPvkJ/ZL5sXkgWrFirK4MibYUdE3OR2nEpuKz1/q8CdNxNQgthZCXYVLjyoDQftaKuniHHWRnPh2GCUetR2/9HsMAXyUT4/3UHwtQT2AggSCGKeSAsFnxBIOuAggdh3AKTL7pDuCyABcMb0aQP7aM4AnAbc/wHwA5D2wDHTTe56gIIOUA/4YYV2e1sg713PXdZJAuncdZMAGkAukU9OAn40O849+0ornPwT93rphWF0mgAbauUrEOthlX8Zu7P5A6kZyKCJy75hhw1Mgr9RAUvX7A3csGqZegEdniCx30c3agAAAABJRU5ErkJggg==" alt="gensokyo">
  </a>
  <a href="https://github.com/hoshinonyaruko/gensokyo-dashboard/actions">
    <img src="images/badge.svg" alt="action">
  </a>
  <a href="https://goreportcard.com/report/github.com/hoshinonyaruko/gensokyo-dashboard">
  <img src="https://goreportcard.com/badge/github.com/hoshinonyaruko/gensokyo-dashboard" alt="GoReportCard">
  </a>
</p>

<p align="center">
  <a href="https://github.com/howmanybots/onebot/blob/master/README.md">文档</a>
  ·
  <a href="https://github.com/hoshinonyaruko/gensokyo-dashboard/releases">下载</a>
  ·
  <a href="https://github.com/hoshinonyaruko/gensokyo-dashboard/releases">开始使用</a>
  ·
  <a href="https://github.com/hoshinonyaruko/gensokyo-dashboard/blob/master/CONTRIBUTING.md">参与贡献</a>
</p>
<p align="center">
  <a href="https://gensokyo.bot">项目主页:gensokyo.bot</a>
</p>

## 介绍

独立的机器人数据统计框架,不再漫无目标,让数据与思考驱动高质量运营。

gensokyo-dashboard是为运营多个机器人的im机器人运营者打造的状态和数据监测工具,随时查看多个机器人在线和发送状态,可以跟踪多项指标,助你提高dau和运营质量.

专为多个机器人场景设计,多个机器人不再需要多开插件,这款控制台与渲染图片,发给用户看的统计插件不同,面向的是注重数据分析的im机器人开发和运营。

## 帮助与支持
持续完善中.....交流群:196173384

欢迎测试,询问任何有关使用的问题,有问必答,有难必帮~

## 兼容性与用法

下载,双击运行,进入webui地址,在web面板,配置您的机器人

可连接多种,支持使用onebotv11协议的全部机器人,支持QQ开放平台机器人填写token和client_secret(app_secret)直接连接,后续还会支持satori和其他bot类型的连接

## WEBUI地址

启动后默认开放到18630端口,地址http://127.0.0.1:18630/webui

可开放18630并访问,有设计登录系统。

## 使用场景

开放sql数据库权限,可二次开发功能

机器人日dau统计

机器人在线情况统计

机器人每日收发统计

机器人指令频次统计

机器人收信息日志储存

用户每日调用次数统计

用户喜好指令分析统计

群每日调用次数统计

数据区分机器人和日期,储存在sql中,可分析群调用变化趋势\用户调用次数变化趋势等重要数据.

## 关于 ISSUE

以下 ISSUE 会被直接关闭

- 提交 BUG 不使用 Template
- 询问已知问题
- 提问找不到重点
- 重复提问

> 请注意, 开发者并没有义务回复您的问题. 您应该具备基本的提问技巧。  
> 有关如何提问，请阅读[《提问的智慧》](https://github.com/ryanhanwu/How-To-Ask-Questions-The-Smart-Way/blob/main/README-zh_CN.md)

## 性能

1mb内存占用 端口错开可多开 稳定运行无报错 连续任务不崩溃中断 可断点续发