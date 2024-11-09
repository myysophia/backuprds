// 使用全局声明的组件
const { Layout, Header, Content, Typography, Title } = window.antdComponents;

function App() {
    console.log('App component rendering'); // 添加日志
    return (
        <Layout>
            <Header style={{ 
                background: '#fff', 
                padding: '0 20px',
                textAlign: 'center'
            }}>
                <Title 
                    level={3} 
                    style={{ 
                        margin: '16px 0',
                        fontWeight: 'bold'
                    }}
                >
                    Nova RDS 跨云灾备系统
                </Title>
            </Header>
            <Content style={{ padding: '20px' }}>
                <BackupList />
            </Content>
        </Layout>
    );
}

// 修改渲染逻辑
const renderApp = () => {
    console.log('Starting to render app'); // 添加日志
    const container = document.getElementById('root');
    if (container) {
        const root = ReactDOM.createRoot(container);
        root.render(React.createElement(App));
    } else {
        console.error('Root element not found');
    }
};

// 确保在 DOM 和所有脚本加载完成后执行
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', renderApp);
} else {
    renderApp();
}
