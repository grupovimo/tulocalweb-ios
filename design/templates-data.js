(function () {
  const U = id => /^https?:/.test(id) ? id : `https://images.unsplash.com/photo-${id}?auto=format&fit=crop&w=700&q=72`;
  const SERIF = "Georgia, 'Times New Roman', serif", SANS = "'Instrument Sans', system-ui, sans-serif";
  const T = [
    ['ceibo', 'Ceibo', 'Moda', 'ed', 'Tienda Lumbre', '1573612664822-d7d347da7b80', '#F4EFEA', '#1C1A17', '#1C1A17', SERIF, '2px', 'Nueva temporada'],
    ['mostaza', 'Mostaza', 'Gastronomía', 'grid', 'Café Bruma', '1495474472287-4d71bcdd2085', '#FBF3E4', '#3B2416', '#C9821A', SANS, '14px', 'Tostado esta semana'],
    ['yerbal', 'Yerbal', 'Almacén', 'cat', 'Almacén Verde', 'https://tulocalweb.com.ar/demo/almacenverde/portada.webp', '#EEF1E8', '#1E2B22', '#2F5D50', SANS, '10px', 'Orgánico y a granel'],
    ['pampa', 'Pampa', 'Moda', 'grid', 'PMPA Store', '1542291026-7eec264c27ff', '#141414', '#FFFFFF', '#E6FF4D', SANS, '6px', 'Drop 07 · ya disponible'],
    ['rosedal', 'Rosedal', 'Belleza', 'ed', 'Rosa Mosqueta', '1612817288484-6f916006741a', '#F8EDEB', '#3A1F25', '#C26A7A', SERIF, '18px', 'Rutina de otoño'],
    ['quebracho', 'Quebracho', 'Hogar', 'ed', 'Casa Nido', '1555041469-a586c61ea9bc', '#EFE9E1', '#2B2520', '#8A5A3E', SERIF, '4px', 'Muebles hechos a mano'],
    ['andes', 'Andes', 'Tecnología', 'grid', 'Sonido Sur', '1505740420928-5e560c06d30e', '#F2F4F7', '#0F1720', '#2563EB', SANS, '16px', 'Envío en 24 h'],
    ['malbec', 'Malbec', 'Bebidas', 'ed', 'Vinoteca Tres', '1506377247377-2a5b3b417ebb', '#1E0E12', '#F4E9E1', '#C24A5A', SERIF, '2px', 'Selección del mes'],
    ['patio', 'Patio', 'Mascotas', 'grid', 'Huellitas', '1543466835-00a7907e9de1', '#FFF7E8', '#2A1E10', '#F08A24', SANS, '22px', 'Todo para tu compañero'],
    ['miga', 'Miga', 'Gastronomía', 'cat', 'Panadería Lola', '1509440159596-0249088772ff', '#FBF1E6', '#3A2414', '#B5651D', SERIF, '12px', 'Masa madre todos los días'],
    ['taller', 'Taller', 'Artesanías', 'ed', 'Madera Viva', '1687422810663-c316494f725a', '#F3EEE6', '#2A231C', '#6B4A2F', SANS, '8px', 'Piezas únicas'],
    ['hora', 'Hora', 'Accesorios', 'grid', 'Relojería Norte', '1522312346375-d1a52e2b99b3', '#F5F5F2', '#161616', '#B08D57', SERIF, '0px', 'Clásicos que duran'],
    ['litoral', 'Litoral', 'Regionales', 'cat', 'Mate & Co', '1687920492018-e21cad9023d9', '#F3E9DC', '#3B2416', '#6B3A22', SANS, '14px', 'El mate de cada mañana'],
    ['torno', 'Torno', 'Arte y diseño', 'ed', 'Estudio Arcilla', '1493106641515-6b5631de4bb9', '#EDE6DD', '#2C2621', '#A4623D', SERIF, '24px', 'Cerámica de autor'],
    ['tijera', 'Tijera', 'Servicios', 'grid', 'Navaja', '1503951914875-452162b0f3f1', '#F6F4F1', '#1E1B2E', '#5B3FD6', SANS, '18px', 'Reservá tu turno online'],
    ['lente', 'Lente', 'Accesorios', 'grid', 'Óptica Faro', '1572635196237-14b3f281503f', '#FFF8EC', '#1F1A12', '#E0A31B', SANS, '12px', '2×1 en anteojos de sol']
  ];
  const POS = ['20% 30%', '80% 40%', '50% 80%', '30% 70%', '70% 20%', '50% 50%'];
  window.MOSTRADOR_TEMPLATES = T.map(([id, name, rubro, layout, store, img, bg, ink, acc, ff, r, tag], i) => ({
    id, name, rubro, layout, store, img: U(img), bg, ink, acc, ff, r, tag,
    dark: ['pampa', 'malbec'].includes(id),
    demo: { ceibo: 'Plantilla Ceibo.dc.html', mostaza: 'Plantilla Mostaza.dc.html', yerbal: 'Plantilla Yerbal.dc.html', tijera: 'Plantilla Tijera.dc.html', pampa: 'Plantilla Pampa.dc.html', malbec: 'Plantilla Malbec.dc.html', miga: 'Plantilla Miga.dc.html', rosedal: 'Plantilla Rosedal.dc.html', andes: 'Plantilla Andes.dc.html', quebracho: 'Plantilla Quebracho.dc.html' }[id] || ('Plantillas Tienda.dc.html?t=' + id),
    prods: POS.map((p, j) => ({ pos: p, price: '$ ' + (8 + ((i * 7 + j * 13) % 40)) + '.900' }))
  }));
  window.MOSTRADOR_RUBROS = ['Todos', 'Moda', 'Gastronomía', 'Almacén', 'Belleza', 'Hogar', 'Tecnología', 'Bebidas', 'Mascotas', 'Artesanías', 'Accesorios', 'Regionales', 'Arte y diseño', 'Servicios'];
})();
