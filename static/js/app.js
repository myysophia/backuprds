const { Layout, Header, Content, Typography } = window.antdComponents;
const { Title } = Typography;

function App() {
    console.log('App component rendering'); // 添加日志
    return (
        <Layout>
            <Header className="site-header">
                <Title 
                    level={3} 
                    className="site-title"
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
        root.render(<App />);  // 使用 JSX
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
