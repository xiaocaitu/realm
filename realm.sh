#!/bin/bash

# 定义颜色变量
red="\033[0;31m"
green="\033[0;32m"
plain="\033[0m"

# 脚本版本
sh_ver="1.1"

# 配置文件路径
CONFIG_PATH="/root/realm/config.toml"

# 初始化环境目录
init_env() {
    mkdir -p /root/realm
}

# 更新脚本
Update_Shell() {
    echo -e "当前版本为 [ ${sh_ver} ]，开始检测最新版本..."
    sh_new_ver=$(wget --no-check-certificate -qO- "https://raw.githubusercontent.com/wcwq98/realm/main/realm.sh" | grep 'sh_ver="' | awk -F "=" '{print $NF}' | sed 's/\"//g' | head -1)
    [[ -z ${sh_new_ver} ]] && echo -e "${red}检测最新版本失败！${plain}" && return
    if [[ ${sh_new_ver} != ${sh_ver} ]]; then
        echo -e "发现新版本 [ ${sh_new_ver} ]，是否更新？[Y/n]"
        read -p "(默认: y):" yn
        [[ -z "${yn}" ]] && yn="y"
        if [[ ${yn} == [Yy] ]]; then
            wget -N --no-check-certificate https://raw.githubusercontent.com/wcwq98/realm/main/realm.sh && chmod +x realm.sh
            echo -e "脚本已更新为最新版本 [ ${sh_new_ver} ]！"
            exit 0
        else
            echo -e "已取消更新。"
        fi
    else
        echo -e "当前已是最新版本 [ ${sh_new_ver} ]！"
    fi
}

# 初始化realm状态
update_realm_status() {
    if [ -f "/root/realm/realm" ]; then
        realm_status="已安装"
        realm_status_color=$green
    else
        realm_status="未安装"
        realm_status_color=$red
    fi
}

# 检查realm服务状态
check_realm_service_status() {
    if systemctl is-active --quiet realm; then
        echo -e "${green}启用${plain}"
    else
        echo -e "${red}未启用${plain}"
    fi
}

# 显示菜单的函数
show_menu() {
    clear
    echo "欢迎使用realm一键转发脚本"
    echo "================="
    echo "1. 部署环境"
    echo "2. 添加转发"
    echo "3. 添加端口段转发"
    echo "4. 删除转发"
    echo "5. 启动服务"
    echo "6. 停止服务"
    echo "7. 重启服务"
    echo "8. 检测更新"
    echo "9. 一键卸载"
    echo "10. 更新脚本"
    echo "0. 退出脚本"
    echo "================="
    echo -e "realm 状态：${realm_status_color}${realm_status}${plain}"
    echo -n "realm 转发状态："
    check_realm_service_status
}

# 部署环境的函数
deploy_realm() {
    init_env
    cd /root/realm

    _version=$(curl -s https://api.github.com/repos/zhboner/realm/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$_version" ]; then
        echo "获取版本号失败，请检查网络连接。"
        return 1
    else
        echo "当前最新版本为: ${_version}"
    fi

    arch=$(uname -m)
    os=$(uname -s | tr '[:upper:]' '[:lower:]')

    case "$arch-$os" in
        x86_64-linux)
            download_url="https://github.com/zhboner/realm/releases/download/${_version}/realm-x86_64-unknown-linux-gnu.tar.gz"
            ;;
        *)
            echo "不支持的架构: $arch-$os"
            return
            ;;
    esac

    wget -O "realm-${_version}.tar.gz" "$download_url"
    tar -xvf "realm-${_version}.tar.gz"
    chmod +x realm

    # 创建默认配置文件
    cat <<EOF > $CONFIG_PATH

[network]
no_tcp = false #是否关闭tcp转发
use_udp = true #是否开启udp转发

#参考模板
# [[endpoints]]
# listen = "0.0.0.0:本地端口"
# remote = "落地鸡ip:目标端口"

[[endpoints]]
listen = "0.0.0.0:1234"
remote = "0.0.0.0:5678"

EOF

    echo "[Unit]
Description=realm
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
Restart=on-failure
RestartSec=5s
WorkingDirectory=/root/realm
ExecStart=/root/realm/realm -c $CONFIG_PATH

[Install]
WantedBy=multi-user.target" > /etc/systemd/system/realm.service

    systemctl daemon-reload
    update_realm_status
    echo "部署完成。"
}

# 初始化realm状态
update_realm_status

# 主循环
while true; do
    show_menu
    read -p "请选择一个选项: " choice
    case $choice in
        1) deploy_realm ;;
        2) add_forward ;;
        3) add_port_range_forward ;;
        4) delete_forward ;;
        5) start_service ;;
        6) stop_service ;;
        7) restart_service ;;
        8) update_realm ;;
        9) uninstall_realm ;;
        10) Update_Shell ;;
        0) echo "退出脚本。"; exit 0 ;;
        *) echo "无效选项: $choice" ;;
    esac
    read -p "按任意键继续..." key
done
