console.log('App component loading...');

// 从全局对象中获取所需的组件
const {
    Layout,
    Typography,
    Card,
    Button
} = window.antComponents;

const {
    ApiOutlined
} = window.antIcons;

// 解构需要的组件
const { Header, Content } = Layout;
const { Title } = Typography;

// 错误边界组件
class ErrorBoundary extends React.Component {
    constructor(props) {
        super(props);
        this.state = { hasError: false, error: null };
    }

    static getDerivedStateFromError(error) {
        return { hasError: true, error };
    }

    componentDidCatch(error, errorInfo) {
        console.error('App error:', error, errorInfo);
    }

    render() {
        if (this.state.hasError) {
            return <div>应用加载错误: {this.state.error.message}</div>;
        }
        return this.props.children;
    }
}

function App() {
    console.log('Rendering App component');
    
    try {
        return (
            <Layout>
                <Header className="site-header">
                    <Typography.Title level={3} className="site-title">
                        Nova RDS 跨云灾备系统
                    </Typography.Title>
                </Header>
                <Layout.Content className="site-content">
                    <div>
                        <BackupList />
                    </div>
                </Layout.Content>
            </Layout>
        );
    } catch (error) {
        console.error('Error in App render:', error);
        return <div>Error rendering App: {error.message}</div>;
    }
}

// 添加重试计数和最大重试次数
let retryCount = 0;
const MAX_RETRIES = 10;

// 修改初始化函数
const initApp = () => {
    console.log(`Initialization attempt ${retryCount + 1}/${MAX_RETRIES}`);
    
    // 详细打印依赖状态
    const dependencies = {
        antComponents: window.antComponents,
        antIcons: window.antIcons,
        React: window.React,
        ReactDOM: window.ReactDOM,
        BackupList: window.BackupList
    };
    
    const missingDeps = Object.entries(dependencies)
        .filter(([key, value]) => !value)
        .map(([key]) => key);

    if (missingDeps.length === 0) {
        console.log('All dependencies loaded, rendering app...');
        renderApp();
        return;
    }

    console.error('Missing dependencies:', missingDeps);
    
    // 添加重试限制
    if (retryCount < MAX_RETRIES) {
        retryCount++;
        setTimeout(initApp, 100);
    } else {
        console.error('Failed to initialize after maximum retries');
        document.getElementById('app').innerHTML = 
            `<div style="color: red; padding: 20px;">
                Failed to load application. Missing dependencies: ${missingDeps.join(', ')}
            </div>`;
    }
};

// 修改渲染函数以确保只执行一次
let hasRendered = false;
const renderApp = () => {
    if (hasRendered) {
        console.log('App already rendered, skipping...');
        return;
    }
    
    console.log('Starting to render app');
    const container = document.getElementById('app');
    
    try {
        ReactDOM.render(
            <ErrorBoundary>
                <App />
            </ErrorBoundary>,
            container,
            () => {
                console.log('App rendered successfully');
                hasRendered = true;
            }
        );
    } catch (error) {
        console.error('Error in renderApp:', error);
        container.innerHTML = `<div>Failed to render app: ${error.message}</div>`;
    }
};

// 根据文档加载状态决定何时初始化应用
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        console.log('DOM Content Loaded');
        initApp();
    });
} else {
    console.log('DOM already loaded');
    initApp();
}

// 为了以防万一，也添加一个 window.onload 处理
window.addEventListener('load', () => {
    console.log('Window loaded');
    initApp();
});

console.log('App component loaded');