-- Migration 012: Seed exercises library with common exercises

INSERT INTO exercises (id, name, description, thumbnail_url, muscles, instructions, tips, difficulty, equipment, video_url)
VALUES
-- PEITO (Chest)
(gen_random_uuid(), 'Supino Reto com Barra', 'Exercício clássico para desenvolvimento do peitoral maior.', '/assets/exercises/supino-reto.png', '["Peito","Tríceps","Ombros"]',
 'Deite no banco, agarre a barra com pegada levemente maior que a largura dos ombros, desça até tocar o peito e empurre de volta ao ponto inicial.',
 'Mantenha os pés firmes no chão. Não deixe os cotovelos abrirem demais.', 'Intermediário', 'Barra', NULL),

(gen_random_uuid(), 'Supino Inclinado com Halteres', 'Variação do supino que enfatiza a porção superior do peitoral.', '/assets/exercises/supino-inclinado.png', '["Peito","Tríceps","Ombros"]',
 'Posicione o banco em 30-45 graus. Segure os halteres na altura do peito, empurre para cima e junte levemente no topo.',
 'Evite inclinar demais o banco (acima de 45 graus enfatiza mais os ombros).', 'Intermediário', 'Halteres', NULL),

(gen_random_uuid(), 'Supino Declinado com Barra', 'Exercício que foca na porção inferior do peitoral.', '/assets/exercises/supino-declinado.png', '["Peito","Tríceps"]',
 'Deite no banco declinado com os pés presos. Agarre a barra, desça até o peito inferior e empurre de volta.',
 'Utilize um spotter por segurança. Controle bem a descida.', 'Intermediário', 'Barra', NULL),

(gen_random_uuid(), 'Crucifixo com Halteres', 'Exercício de isolamento para o peitoral.', '/assets/exercises/crucifixo.png', '["Peito"]',
 'Deite no banco, segure os halteres acima do peito com cotovelos levemente flexionados. Abra os braços em arco até sentir o alongamento no peito, retorne.',
 'Foco no alongamento e contração. Use cargas moderadas para proteger os ombros.', 'Iniciante', 'Halteres', NULL),

(gen_random_uuid(), 'Mergulho no Paralelo', 'Exercício de peso corporal para peitoral e tríceps.', '/assets/exercises/mergulho-paralelo.png', '["Peito","Tríceps","Ombros"]',
 'Segure as barras paralelas, desça o corpo flexionando os cotovelos até 90 graus, mantenha o tronco levemente inclinado à frente, retorne.',
 'Incline o corpo à frente para enfatizar o peitoral. Corpo reto enfatiza o tríceps.', 'Intermediário', 'Peso corporal', NULL),

-- COSTAS (Back)
(gen_random_uuid(), 'Puxada Frontal na Máquina', 'Exercício fundamental para desenvolvimento das costas.', '/assets/exercises/puxada-frontal.png', '["Costas","Bíceps"]',
 'Sente na máquina, segure a barra com pegada pronada larga. Puxe a barra até a altura do queixo contraindo as costas, retorne controlado.',
 'Não balance o tronco. Foque em puxar com os cotovelos, não com as mãos.', 'Iniciante', 'Máquina', NULL),

(gen_random_uuid(), 'Remada Curvada com Barra', 'Exercício composto para espessura das costas.', '/assets/exercises/remada-curvada.png', '["Costas","Bíceps","Lombar"]',
 'Incline o tronco a 45 graus, segure a barra com pegada pronada. Puxe a barra em direção ao abdômen mantendo as costas retas.',
 'Mantenha a coluna neutra. Não use impulso excessivo do tronco.', 'Intermediário', 'Barra', NULL),

(gen_random_uuid(), 'Pull-up (Barra Fixa)', 'Exercício de peso corporal para costas e bíceps.', '/assets/exercises/pull-up.png', '["Costas","Bíceps"]',
 'Segure a barra com pegada pronada, braços estendidos. Puxe o corpo até o queixo ultrapassar a barra, desça controlado.',
 'Evite balançar o corpo. Se necessário, use elástico de resistência para assistência.', 'Avançado', 'Peso corporal', NULL),

(gen_random_uuid(), 'Remada Unilateral com Haltere', 'Exercício unilateral para costas com apoio no banco.', '/assets/exercises/remada-unilateral.png', '["Costas","Bíceps"]',
 'Apoie um joelho e a mão no banco. Segure o haltere com a outra mão e puxe em direção ao quadril mantendo o cotovelo próximo ao corpo.',
 'Mantenha as costas paralelas ao chão. Controle o movimento nos dois sentidos.', 'Iniciante', 'Halteres', NULL),

(gen_random_uuid(), 'Levantamento Terra', 'Exercício composto que trabalha toda a cadeia posterior.', '/assets/exercises/levantamento-terra.png', '["Costas","Glúteos","Pernas","Lombar"]',
 'Pé na largura dos ombros, barra sobre os pés. Agache e segure a barra com pegada mista. Empurre o chão, estenda quadril e joelhos até ficar em pé.',
 'A coluna DEVE permanecer neutra durante todo o movimento. Não arredonde as costas.', 'Avançado', 'Barra', NULL),

-- PERNAS (Legs)
(gen_random_uuid(), 'Agachamento com Barra', 'Exercício rainha para desenvolvimento das pernas.', '/assets/exercises/agachamento.png', '["Quadríceps","Glúteos","Isquiotibiais"]',
 'Barra sobre os trapézios, pés na largura dos ombros. Desça como se fosse sentar, mantenha o peito alto, desça até as coxas ficarem paralelas ao chão.',
 'Joelhos alinhados com os pés. Não deixe os joelhos caírem para dentro.', 'Intermediário', 'Barra', NULL),

(gen_random_uuid(), 'Leg Press 45°', 'Exercício de musculação para quadríceps e glúteos.', '/assets/exercises/leg-press.png', '["Quadríceps","Glúteos","Isquiotibiais"]',
 'Sente no aparelho, posicione os pés na plataforma na largura dos ombros. Desça controlado até 90 graus e empurre de volta sem travar os joelhos.',
 'Não bloqueie os joelhos na extensão. Ajuste a posição dos pés para variar o foco muscular.', 'Iniciante', 'Máquina', NULL),

(gen_random_uuid(), 'Cadeira Extensora', 'Exercício de isolamento para o quadríceps.', '/assets/exercises/cadeira-extensora.png', '["Quadríceps"]',
 'Sente na máquina com os pés sob o apoio. Estenda as pernas completamente e desça de forma controlada.',
 'Faça uma pausa de 1 segundo na extensão total para maximizar a contração.', 'Iniciante', 'Máquina', NULL),

(gen_random_uuid(), 'Mesa Flexora', 'Exercício de isolamento para os isquiotibiais.', '/assets/exercises/mesa-flexora.png', '["Isquiotibiais"]',
 'Deite na máquina com os tornozelos sob o apoio. Flexione os joelhos puxando os calcanhares em direção aos glúteos, retorne controlado.',
 'Mantenha o quadril pressionado contra o banco durante todo o movimento.', 'Iniciante', 'Máquina', NULL),

(gen_random_uuid(), 'Afundo com Halteres', 'Exercício unilateral para pernas e glúteos.', '/assets/exercises/afundo.png', '["Quadríceps","Glúteos","Isquiotibiais"]',
 'Fique em pé com halteres nas mãos. Dê um passo à frente, desça até o joelho traseiro quase tocar o chão, empurre e retorne à posição inicial.',
 'Mantenha o torso ereto e o joelho dianteiro alinhado com o pé.', 'Intermediário', 'Halteres', NULL),

(gen_random_uuid(), 'Elevação de Panturrilha em Pé', 'Exercício para desenvolvimento do músculo gastrocnêmio.', '/assets/exercises/elevacao-panturrilha.png', '["Panturrilha"]',
 'Fique em pé com a ponta dos pés na borda de um step. Desça os calcanhares abaixo do nível do step, suba na ponta dos pés o máximo possível.',
 'Realize o movimento em amplitude completa. Pause 1 segundo no topo.', 'Iniciante', 'Peso corporal', NULL),

-- OMBROS (Shoulders)
(gen_random_uuid(), 'Desenvolvimento com Barra', 'Exercício composto para ombros.', '/assets/exercises/desenvolvimento-barra.png', '["Ombros","Tríceps"]',
 'Sentado ou em pé, segure a barra na frente com pegada pronada na largura dos ombros. Empurre acima da cabeça até os braços estendidos, desça controlado.',
 'Não arquee a lombar. Mantenha o core contraído durante todo o movimento.', 'Intermediário', 'Barra', NULL),

(gen_random_uuid(), 'Elevação Lateral com Halteres', 'Exercício de isolamento para o deltoide lateral.', '/assets/exercises/elevacao-lateral.png', '["Ombros"]',
 'Fique em pé com halteres nas laterais do corpo. Eleve os braços lateralmente até a altura dos ombros com cotovelos levemente flexionados, retorne.',
 'Não balance o corpo. Use cargas moderadas para manter a técnica.', 'Iniciante', 'Halteres', NULL),

(gen_random_uuid(), 'Elevação Frontal com Halteres', 'Exercício para o deltoide anterior.', '/assets/exercises/elevacao-frontal.png', '["Ombros"]',
 'Fique em pé com halteres à frente das coxas. Eleve um braço de cada vez até a altura dos ombros, retorne e alterne.',
 'Mantenha o movimento controlado e evite usar o impulso do tronco.', 'Iniciante', 'Halteres', NULL),

(gen_random_uuid(), 'Remada Alta com Barra', 'Exercício para trapézio e deltoides.', '/assets/exercises/remada-alta.png', '["Ombros","Trapézio"]',
 'Segure a barra com pegada pronada estreita. Puxe a barra verticalmente até a altura do queixo, cotovelos acima dos punhos, retorne.',
 'Mantenha a barra próxima ao corpo. Evite pegadas muito estreitas para proteger os ombros.', 'Intermediário', 'Barra', NULL),

-- BRAÇOS (Arms)
(gen_random_uuid(), 'Rosca Direta com Barra', 'Exercício clássico para bíceps.', '/assets/exercises/rosca-direta.png', '["Bíceps"]',
 'Fique em pé com a barra, pegada supinada na largura dos ombros, cotovelos junto ao corpo. Flexione os cotovelos levando a barra até os ombros, desça controlado.',
 'Mantenha os cotovelos fixos. Não balance o tronco para dar impulso.', 'Iniciante', 'Barra', NULL),

(gen_random_uuid(), 'Rosca Alternada com Halteres', 'Exercício unilateral para bíceps.', '/assets/exercises/rosca-alternada.png', '["Bíceps"]',
 'Sentado ou em pé com halteres nos lados. Flexione um braço de cada vez rodando o punho supinando no topo do movimento.',
 'A supinação do punho aumenta o pico de contração do bíceps.', 'Iniciante', 'Halteres', NULL),

(gen_random_uuid(), 'Tríceps Pulley com Corda', 'Exercício de isolamento para tríceps no cabo.', '/assets/exercises/triceps-corda.png', '["Tríceps"]',
 'Posicione-se em frente ao cabo com a corda em cima. Segure a corda com ambas as mãos, cotovelos junto ao corpo. Estenda os cotovelos para baixo e abra a corda no final.',
 'Mantenha os cotovelos fixos durante todo o movimento.', 'Iniciante', 'Cabo', NULL),

(gen_random_uuid(), 'Tríceps Testa com Halteres', 'Exercício para tríceps deitado no banco.', '/assets/exercises/triceps-testa.png', '["Tríceps"]',
 'Deite no banco com halteres acima do peito. Flexione os cotovelos abaixando os halteres em direção à testa, estenda de volta.',
 'Mantenha os cotovelos apontados para cima durante todo o movimento.', 'Intermediário', 'Halteres', NULL),

(gen_random_uuid(), 'Rosca Concentrada', 'Exercício de isolamento para bíceps com apoio no joelho.', '/assets/exercises/rosca-concentrada.png', '["Bíceps"]',
 'Sentado, apoie o cotovelo na parte interna da coxa. Flexione completamente o braço, retorne controlado.',
 'Movimento muito eficaz para isolamento do bíceps. Varie a supinação no topo.', 'Iniciante', 'Halteres', NULL),

-- CORE
(gen_random_uuid(), 'Prancha Abdominal', 'Exercício isométrico fundamental para o core.', '/assets/exercises/prancha.png', '["Core","Lombar","Ombros"]',
 'Apoie os antebraços e a ponta dos pés no chão. Mantenha o corpo em linha reta da cabeça aos pés, contraindo o abdômen.',
 'Não deixe o quadril cair ou subir. Respire normalmente durante a isometria.', 'Iniciante', 'Peso corporal', NULL),

(gen_random_uuid(), 'Abdominal Crunch', 'Exercício básico para o reto abdominal.', '/assets/exercises/crunch.png', '["Core"]',
 'Deite de costas com joelhos flexionados, mãos atrás da cabeça. Contraia o abdômen levantando apenas os ombros do chão, desça controlado.',
 'Não puxe o pescoço com as mãos. O movimento é curto e controlado.', 'Iniciante', 'Peso corporal', NULL),

(gen_random_uuid(), 'Russian Twist', 'Exercício rotacional para o core.', '/assets/exercises/russian-twist.png', '["Core","Oblíquos"]',
 'Sentado com joelhos flexionados e pés levemente elevados. Segure um peso ou apenas as mãos juntas. Gire o tronco de lado a lado.',
 'Mantenha as costas levemente inclinadas. Aumente a dificuldade elevando mais os pés.', 'Iniciante', 'Peso corporal', NULL),

(gen_random_uuid(), 'Elevação de Pernas', 'Exercício para abdômen inferior.', '/assets/exercises/elevacao-pernas.png', '["Core"]',
 'Deite de costas com as mãos sob o glúteo. Eleve as pernas juntas até 90 graus, desça controlado sem tocar o chão.',
 'Mantenha a lombar pressionada contra o chão. Não deixe as pernas caírem bruscamente.', 'Intermediário', 'Peso corporal', NULL),

(gen_random_uuid(), 'Superman', 'Exercício para fortalecimento da lombar.', '/assets/exercises/superman.png', '["Lombar","Glúteos"]',
 'Deite de bruços com braços estendidos à frente. Eleve simultaneamente os braços e pernas do chão, segure 2 segundos, desça.',
 'Mantenha o pescoço neutro. Evite hiperestender a coluna cervical.', 'Iniciante', 'Peso corporal', NULL)

ON CONFLICT DO NOTHING;
