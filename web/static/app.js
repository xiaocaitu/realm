document.addEventListener('DOMContentLoaded', () => {
    const outputDiv = document.getElementById('output');
    const startButton = document.getElementById('startButton');
    const stopButton = document.getElementById('stopButton');
    const addRuleButton = document.getElementById('addRuleButton');
    const localPortInput = document.getElementById('localPort');
    const remoteIPInput = document.getElementById('remoteIP');
    const remotePortInput = document.getElementById('remotePort');

    async function fetchForwardingRules() {
        try {
            const response = await fetch('/get_rules');
            if (!response.ok) {
                throw new Error('获取规则失败：' + response.statusText);
            }
            const rules = await response.json();
            const tbody = document.querySelector('#forwardingTable tbody');
            tbody.innerHTML = '';

            rules.forEach((rule, index) => {
                const [, localPort] = rule.Listen.split(':');
                const lastColonIndex = rule.Remote.lastIndexOf(':');
                const remoteIP = rule.Remote.substring(0, lastColonIndex);
                const remotePort = rule.Remote.substring(lastColonIndex + 1);

                const row = document.createElement('tr');
                row.innerHTML = `
                    <td>${index + 1}</td>
                    <td>${localPort}</td>
                    <td>${remoteIP}</td>
                    <td>${remotePort}</td>
                    <td><button data-listen="${rule.Listen}" class="delete-btn">删除</button></td>
                `;
                tbody.appendChild(row);
            });

            // 为删除按钮添加事件监听
            const deleteButtons = document.querySelectorAll('.delete-btn');
            deleteButtons.forEach(button => {
                button.addEventListener('click', function() {
                    const listenAddress = this.getAttribute('data-listen');
                    deleteRule(listenAddress);
                });
            });
        } catch (error) {
            console.error('请求失败：', error);
            outputDiv.textContent = '获取转发规则失败';
        }
    }

    async function deleteRule(listenAddress) {
        try {
            const response = await fetch(`/delete_rule?listen=${encodeURIComponent(listenAddress)}`, {
                method: 'DELETE'
            });
            if (!response.ok) {
                throw new Error('删除规则失败：' + response.statusText);
            }
            fetchForwardingRules(); // 重新获取规则列表
        } catch (error) {
            console.error('删除规则失败：', error);
            outputDiv.textContent = '删除规则失败';
        }
    }

    async function addRule() {
        const localPort = localPortInput.value;
        const remoteIP = remoteIPInput.value;
        const remotePort = remotePortInput.value;

        if (!localPort || !remoteIP || !remotePort) {
            outputDiv.textContent = '请填写所有字段';
            return;
        }

        try {
            const response = await fetch('/add_rule', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    localPort,
                    remoteIP,
                    remotePort
                })
            });
            if (!response.ok) {
                throw new Error('添加规则失败：' + response.statusText);
            }
            fetchForwardingRules(); // 重新获取规则列表
        } catch (error) {
            console.error('添加规则失败：', error);
            outputDiv.textContent = '添加规则失败';
        }
    }

    startButton.addEventListener('click', async () => {
        try {
            const response = await fetch('/start_service', {
                method: 'POST'
            });
            if (!response.ok) {
                throw new Error('启动服务失败：' + response.statusText);
            }
            const result = await response.json();
            outputDiv.textContent = result.output;
        } catch (error) {
            console.error('启动服务失败：', error);
            outputDiv.textContent = '启动服务失败';
        }
    });

    stopButton.addEventListener('click', async () => {
        try {
            const response = await fetch('/stop_service', {
                method: 'POST'
            });
            if (!response.ok) {
                throw new Error('停止服务失败：' + response.statusText);
            }
            const result = await response.json();
            outputDiv.textContent = result.output;
        } catch (error) {
            console.error('停止服务失败：', error);
            outputDiv.textContent = '停止服务失败';
        }
    });

    addRuleButton.addEventListener('click', addRule);

    // 初始化时获取规则列表
    fetchForwardingRules();
});
